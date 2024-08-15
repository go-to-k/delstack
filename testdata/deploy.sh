#!/bin/bash

# This script allows you to deploy the stack for delstack testing.
# Due to quota limitations, only up to [5 test stacks] can be created by this script at the same time.
# Contains [2 AWS::S3Express::DirectoryBucket] : Directory bucket can only have up to 10 buckets created per AWS account (per region).
# Contains [2 AWS::IAM::Group] : 1 IAM user can only belong to 10 IAM groups.  In this script, 1 IAM user is used across multiple script runs.

set -eu

cd $(dirname $0)

profile=""
stage=""
profile_option=""

REGION="us-east-1"

while getopts p:s: OPT; do
	case $OPT in
	p)
		profile="$OPTARG"
		;;
	s)
		stage="$OPTARG"
		;;
	esac
done

if [ -z "${stage}" ]; then
	echo "stage option (-s) is required"
	exit 1
fi

CFN_TEMPLATE="./yamldir/test_root.yaml"
CFN_OUTPUT_TEMPLATE="./yamldir/test_root_output.yaml"

CFN_PJ_PREFIX="dev-${stage}"

CFN_STACK_NAME="${CFN_PJ_PREFIX}-TestStack"

sam_bucket=$(echo "${CFN_STACK_NAME}" | tr '[:upper:]' '[:lower:]')

if [ -n "${profile}" ]; then
	profile_option="--profile ${profile} --region ${REGION}"
fi

account_id=$(aws sts get-caller-identity \
	--query "Account" \
	--output text \
	${profile_option})

### for S3 Buckets
dir="./testfiles/${CFN_STACK_NAME}"
mkdir -p ${dir}
touch ${dir}/{1..1500}.txt

### for ECR
ecr_repository_enddpoint="${account_id}.dkr.ecr.${REGION}.amazonaws.com"
aws ecr get-login-password ${profile_option} |
	docker login --username AWS --password-stdin ${ecr_repository_enddpoint}

image_tag="delstack-test"
docker build -t ${image_tag} .

# The following function is no longer needed as the IAM role no longer fails on normal deletion, but it is left in place just in case.
function attach_policy_to_role() {
	local own_stackname="${1}"

	local attach_policy_arn="arn:aws:iam::${account_id}:policy/DelstackTestPolicy"
	local exists_policy=$(aws iam get-policy \
		--policy-arn "${attach_policy_arn}" \
		--output text \
		${profile_option} 2>/dev/null || :)

	if [ -z "${exists_policy}" ]; then
		aws iam create-policy \
			--policy-name "DelstackTestPolicy" \
			--policy-document file://./policy_document.json \
			--description "test policy" \
			${profile_option} 1>/dev/null
	fi

	local resources=$(
		aws cloudformation list-stack-resources \
			--stack-name ${own_stackname} \
			--query "StackResourceSummaries" \
			${profile_option} |
			jq '.[] | {LogicalResourceId:.LogicalResourceId, PhysicalResourceId:.PhysicalResourceId, ResourceType:.ResourceType}' |
			jq -s '.'
	)

	local iam_role_resources=$(
		echo "${resources}" |
			jq '.[] | select(.ResourceType == "AWS::IAM::Role") | .PhysicalResourceId' |
			jq -s '.'
	)

	local nested_stack_resources=$(
		echo "${resources}" |
			jq '.[] | select(.ResourceType == "AWS::CloudFormation::Stack") | .PhysicalResourceId' |
			jq -s '.'
	)

	local iam_role_resource_len=$(echo $iam_role_resources | jq length)
	local nested_stack_resourceLen=$(echo $nested_stack_resources | jq length)
	local iam_role_name_array=()
	local nested_own_stackname_array=()

	if [ ${iam_role_resource_len} -gt 0 ]; then
		for i in $(seq 0 $(($iam_role_resource_len - 1))); do
			iam_role_name_array+=($(echo $iam_role_resources | jq -r ".[$i]"))
		done
	fi

	if [ ${nested_stack_resourceLen} -gt 0 ]; then
		for i in $(seq 0 $(($nested_stack_resourceLen - 1))); do
			nested_own_stackname_array+=($(
				echo $nested_stack_resources |
					jq -r ".[$i]" |
					sed -e "s/^arn:aws:cloudformation:[^:]*:[0-9]*:stack\/\([^\/]*\)\/.*$/\1/g"
			)
			)
		done

		local pids=()
		for i in ${!nested_own_stackname_array[@]}; do
			attach_policy_to_role "${nested_own_stackname_array[$i]}" &
			pids[$!]=$!
		done
		wait ${pids[@]}
	fi

	for i in ${!iam_role_name_array[@]}; do
		aws iam attach-role-policy \
			--role-name "${iam_role_name_array[$i]}" \
			--policy-arn "${attach_policy_arn}" \
			${profile_option}
	done
}

function attach_user_to_group() {
	local own_stackname="${1}"

	local attach_user_name="DelstackTestUser"
	local exists_user=$(aws iam get-user \
		--user-name "${attach_user_name}" \
		--output text \
		${profile_option} 2>/dev/null || :)

	if [ -z "${exists_user}" ]; then
		aws iam create-user \
			--user-name "${attach_user_name}" \
			${profile_option} 1>/dev/null
	fi

	local resources=$(
		aws cloudformation list-stack-resources \
			--stack-name ${own_stackname} \
			--query "StackResourceSummaries" \
			${profile_option} |
			jq '.[] | {LogicalResourceId:.LogicalResourceId, PhysicalResourceId:.PhysicalResourceId, ResourceType:.ResourceType}' |
			jq -s '.'
	)

	local iam_group_resources=$(
		echo "${resources}" |
			jq '.[] | select(.ResourceType == "AWS::IAM::Group") | .PhysicalResourceId' |
			jq -s '.'
	)

	local nested_stack_resources=$(
		echo "${resources}" |
			jq '.[] | select(.ResourceType == "AWS::CloudFormation::Stack") | .PhysicalResourceId' |
			jq -s '.'
	)

	local iam_group_resource_len=$(echo $iam_group_resources | jq length)
	local nested_stack_resourceLen=$(echo $nested_stack_resources | jq length)
	local iam_group_name_array=()
	local nested_own_stackname_array=()

	if [ ${iam_group_resource_len} -gt 0 ]; then
		for i in $(seq 0 $(($iam_group_resource_len - 1))); do
			iam_group_name_array+=($(echo $iam_group_resources | jq -r ".[$i]"))
		done
	fi

	if [ ${nested_stack_resourceLen} -gt 0 ]; then
		for i in $(seq 0 $(($nested_stack_resourceLen - 1))); do
			nested_own_stackname_array+=($(
				echo $nested_stack_resources |
					jq -r ".[$i]" |
					sed -e "s/^arn:aws:cloudformation:[^:]*:[0-9]*:stack\/\([^\/]*\)\/.*$/\1/g"
			)
			)
		done

		local pids=()
		for i in ${!nested_own_stackname_array[@]}; do
			attach_user_to_group "${nested_own_stackname_array[$i]}" &
			pids[$!]=$!
		done
		wait ${pids[@]}
	fi

	for i in ${!iam_group_name_array[@]}; do
		aws iam add-user-to-group \
			--group-name "${iam_group_name_array[$i]}" \
			--user-name "${attach_user_name}" \
			${profile_option}
	done
}

function build_upload() {
	local own_stackname="${1}"

	local resources=$(
		aws cloudformation list-stack-resources \
			--stack-name ${own_stackname} \
			--query "StackResourceSummaries" \
			${profile_option} |
			jq '.[] | {LogicalResourceId:.LogicalResourceId, PhysicalResourceId:.PhysicalResourceId, ResourceType:.ResourceType}' |
			jq -s '.'
	)

	local ecr_resources=$(
		echo "${resources}" |
			jq '.[] | select(.ResourceType == "AWS::ECR::Repository") | .PhysicalResourceId' |
			jq -s '.'
	)

	local nested_stack_resources=$(
		echo "${resources}" |
			jq '.[] | select(.ResourceType == "AWS::CloudFormation::Stack") | .PhysicalResourceId' |
			jq -s '.'
	)

	local ecr_resource_len=$(echo $ecr_resources | jq length)
	local nested_stack_resourceLen=$(echo $nested_stack_resources | jq length)
	local ecr_name_array=()
	local nested_own_stackname_array=()

	if [ ${ecr_resource_len} -gt 0 ]; then
		for i in $(seq 0 $(($ecr_resource_len - 1))); do
			ecr_name_array+=($(echo $ecr_resources | jq -r ".[$i]"))
		done
	fi

	if [ ${nested_stack_resourceLen} -gt 0 ]; then
		for i in $(seq 0 $(($nested_stack_resourceLen - 1))); do
			nested_own_stackname_array+=($(
				echo $nested_stack_resources |
					jq -r ".[$i]" |
					sed -e "s/^arn:aws:cloudformation:[^:]*:[0-9]*:stack\/\([^\/]*\)\/.*$/\1/g"
			)
			)
		done

		local pids=()
		for i in ${!nested_own_stackname_array[@]}; do
			build_upload "${nested_own_stackname_array[$i]}" &
			pids[$!]=$!
		done
		wait ${pids[@]}
	fi

	for i in ${!ecr_name_array[@]}; do
		local ecr_repository_uri="${ecr_repository_enddpoint}/${ecr_name_array[$i]}"
		local ecr_tag="test"
		local uri_tag="${ecr_repository_uri}:${ecr_tag}"
		docker tag ${image_tag}:latest ${uri_tag}
		docker push ${uri_tag}
	done
}

function object_upload() {
	local own_stackname="${1}"

	local resources=$(
		aws cloudformation list-stack-resources \
			--stack-name ${own_stackname} \
			--query "StackResourceSummaries" \
			${profile_option} |
			jq '.[] | {LogicalResourceId:.LogicalResourceId, PhysicalResourceId:.PhysicalResourceId, ResourceType:.ResourceType}' |
			jq -s '.'
	)

	local bucket_resources=$(
		echo "${resources}" |
			jq '.[] | select(.ResourceType == "AWS::S3::Bucket") | .PhysicalResourceId' |
			jq -s '.'
	)

	local directory_bucket_resources=$(
		echo "${resources}" |
			jq '.[] | select(.ResourceType == "AWS::S3Express::DirectoryBucket") | .PhysicalResourceId' |
			jq -s '.'
	)

	local nested_stack_resources=$(
		echo "${resources}" |
			jq '.[] | select(.ResourceType == "AWS::CloudFormation::Stack") | .PhysicalResourceId' |
			jq -s '.'
	)

	local bucket_resource_len=$(echo $bucket_resources | jq length)
	local directory_bucket_resource_len=$(echo $directory_bucket_resources | jq length)
	local nested_stack_resourceLen=$(echo $nested_stack_resources | jq length)
	local bucket_name_array=()
	local directory_bucket_name_array=()
	local nested_own_stackname_array=()

	if [ ${bucket_resource_len} -gt 0 ]; then
		for i in $(seq 0 $(($bucket_resource_len - 1))); do
			bucket_name_array+=($(echo $bucket_resources | jq -r ".[$i]"))
		done
	fi

	if [ ${directory_bucket_resource_len} -gt 0 ]; then
		for i in $(seq 0 $(($directory_bucket_resource_len - 1))); do
			directory_bucket_name_array+=($(echo $directory_bucket_resources | jq -r ".[$i]"))
		done
	fi

	if [ ${nested_stack_resourceLen} -gt 0 ]; then
		for i in $(seq 0 $(($nested_stack_resourceLen - 1))); do
			nested_own_stackname_array+=($(
				echo $nested_stack_resources |
					jq -r ".[$i]" |
					sed -e "s/^arn:aws:cloudformation:[^:]*:[0-9]*:stack\/\([^\/]*\)\/.*$/\1/g"
			)
			)
		done

		local pids=()
		for i in ${!nested_own_stackname_array[@]}; do
			object_upload "${nested_own_stackname_array[$i]}" &
			pids[$!]=$!
		done
		wait ${pids[@]}
	fi

	for i in ${!bucket_name_array[@]}; do
		aws s3 cp ${dir} s3://${bucket_name_array[$i]}/ --recursive ${profile_option} >/dev/null
		aws s3 cp ${dir} s3://${bucket_name_array[$i]}/ --recursive ${profile_option} >/dev/null # version
		aws s3 rm s3://${bucket_name_array[$i]}/ --recursive ${profile_option} >/dev/null        # delete marker
	done

	for i in ${!directory_bucket_name_array[@]}; do
		# Do not finish even in the event of an error because the following error will occur.
		### upload failed: testfiles/5594.txt to s3://dev-goto-002-descend--use1-az4--x-s3/5594.txt An error occurred (400) when calling the PutObject operation: Bad Request
		set +e
		aws s3 cp ${dir} s3://${directory_bucket_name_array[$i]}/ --recursive ${profile_option} >/dev/null
		set -e
	done
}

function start_backup() {
	local own_stackname="${1}"

	local resources=$(
		aws cloudformation list-stack-resources \
			--stack-name ${own_stackname} \
			--query "StackResourceSummaries" \
			${profile_option} |
			jq '.[] | {LogicalResourceId:.LogicalResourceId, PhysicalResourceId:.PhysicalResourceId, ResourceType:.ResourceType}' |
			jq -s '.'
	)

	local back_vault_resources=$(
		echo "${resources}" |
			jq '.[] | select(.ResourceType == "AWS::Backup::BackupVault") | .PhysicalResourceId' |
			jq -s '.'
	)

	local nested_stack_resources=$(
		echo "${resources}" |
			jq '.[] | select(.ResourceType == "AWS::CloudFormation::Stack") | .PhysicalResourceId' |
			jq -s '.'
	)

	local back_vault_resource_len=$(echo $back_vault_resources | jq length)
	local nested_stack_resourceLen=$(echo $nested_stack_resources | jq length)
	local back_vault_name_array=()
	local nested_own_stackname_array=()

	if [ ${back_vault_resource_len} -gt 0 ]; then
		for i in $(seq 0 $(($back_vault_resource_len - 1))); do
			back_vault_name_array+=($(echo $back_vault_resources | jq -r ".[$i]"))
		done
	fi

	if [ ${nested_stack_resourceLen} -gt 0 ]; then
		for i in $(seq 0 $(($nested_stack_resourceLen - 1))); do
			nested_own_stackname_array+=($(
				echo $nested_stack_resources |
					jq -r ".[$i]" |
					sed -e "s/^arn:aws:cloudformation:[^:]*:[0-9]*:stack\/\([^\/]*\)\/.*$/\1/g"
			)
			)
		done

		local pids=()
		for i in ${!nested_own_stackname_array[@]}; do
			start_backup "${nested_own_stackname_array[$i]}" &
			pids[$!]=$!
		done
		wait ${pids[@]}
	fi

	local resource_arn="arn:aws:dynamodb:${REGION}:${account_id}:table/${CFN_PJ_PREFIX}-Table"
	local iam_role_arn="arn:aws:iam::${account_id}:role/service-role/${CFN_PJ_PREFIX}-AWSBackupServiceRole"
	for i in ${!back_vault_name_array[@]}; do
		local backup_job_id=$(
			aws backup start-backup-job \
				--backup-vault-name "${back_vault_name_array[$i]}" \
				--resource-arn "${resource_arn}" \
				--iam-role-arn "${iam_role_arn}" \
				${profile_option} |
				jq -r '.BackupJobId'
		)

		while true; do
			local state=$(
				aws backup describe-backup-job \
					--backup-job-id "${backup_job_id}" \
					${profile_option} |
					jq -r '.State'
			)
			if [ "${state}" = "COMPLETED" ]; then
				break
			elif [ "${state}" = "FAILED" ] || [ "${state}" = "ABORTED" ]; then
				echo "Backup failed !!"
				exit 1
			fi
			sleep 10
		done
	done
}

if [ -z "$(aws s3 ls ${profile_option} | grep ${sam_bucket})" ]; then
	echo ${profile_option}
	aws s3 mb s3://${sam_bucket} ${profile_option}
fi

sam package \
	--template-file ${CFN_TEMPLATE} \
	--output-template-file ${CFN_OUTPUT_TEMPLATE} \
	--s3-bucket ${sam_bucket} \
	${profile_option}

sam deploy \
	--template-file ${CFN_OUTPUT_TEMPLATE} \
	--stack-name ${CFN_STACK_NAME} \
	--capabilities CAPABILITY_IAM CAPABILITY_AUTO_EXPAND CAPABILITY_NAMED_IAM \
	--parameter-overrides \
	PJPrefix=${CFN_PJ_PREFIX} \
	${profile_option}

echo "=== attach_user_to_group ==="
attach_user_to_group "${CFN_STACK_NAME}"

echo "=== object_upload ==="
object_upload "${CFN_STACK_NAME}"

echo "=== build_upload ==="
build_upload "${CFN_STACK_NAME}"

echo "=== start_backup ==="
start_backup "${CFN_STACK_NAME}"

# The following function is no longer needed as the IAM role no longer fails on normal deletion, but it is left in place just in case.
echo "=== attach_policy_to_role ==="
attach_policy_to_role "${CFN_STACK_NAME}"

rm -rf ${dir}
