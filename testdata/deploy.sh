#!/bin/bash
set -eu

cd `dirname $0`

profile=""
stage=""
profile_option=""

REGION="ap-northeast-1"

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

sam_bucket=`echo "${CFN_STACK_NAME}" | tr '[:upper:]' '[:lower:]'`

if [ -n "${profile}" ]; then
	profile_option="--profile ${profile}"
fi

account_id=$(aws sts get-caller-identity \
	--query "Account" \
	--output text \
	${profile_option})

function attach_policy(){
	local own_stackname="${1}"

	local attach_policy_arn="arn:aws:iam::${account_id}:policy/${CFN_PJ_PREFIX}-TestPolicy"
	local exists_policy=$(aws iam get-policy \
		--policy-arn "${attach_policy_arn}" \
		--output text \
		${profile_option} 2>/dev/null || : )

	if [ -z "${exists_policy}" ]; then
		aws iam create-policy \
			--policy-name "${CFN_PJ_PREFIX}-TestPolicy" \
			--policy-document file://./policy_document.json \
			--description "test policy" \
			${profile_option} 1>/dev/null
	fi

	local resources=$(
		aws cloudformation list-stack-resources \
			--stack-name ${own_stackname} \
			--query "StackResourceSummaries" \
			${profile_option} \
		| jq '.[] | {LogicalResourceId:.LogicalResourceId, PhysicalResourceId:.PhysicalResourceId, ResourceType:.ResourceType}' \
		| jq -s '.'
	)

	local iam_role_resources=$(
		echo "${resources}" \
		| jq '.[] | select(.ResourceType == "AWS::IAM::Role") | .PhysicalResourceId' \
		| jq -s '.'
	)

	local nested_stack_resources=$(
		echo "${resources}" \
		| jq '.[] | select(.ResourceType == "AWS::CloudFormation::Stack") | .PhysicalResourceId' \
		| jq -s '.'
	)

	local iam_role_resource_len=$(echo $iam_role_resources | jq length)
	local nested_stack_resourceLen=$(echo $nested_stack_resources | jq length)
	local iam_role_name_array=()
	local nested_own_stackname_array=()


	if [ ${iam_role_resource_len} -gt 0 ]; then
		for i in $( seq 0 $(($iam_role_resource_len - 1)) ); do 
			iam_role_name_array+=( $(echo $iam_role_resources | jq -r ".[$i]") )
		done
	fi

	if [ ${nested_stack_resourceLen} -gt 0 ]; then
		for i in $( seq 0 $(($nested_stack_resourceLen - 1)) ); do 
			nested_own_stackname_array+=( $(
				echo $nested_stack_resources \
				| jq -r ".[$i]" \
				| sed -e "s/^arn:aws:cloudformation:[^:]*:[0-9]*:stack\/\([^\/]*\)\/.*$/\1/g" \
				) \
			)
		done

		local pids=()
		for i in ${!nested_own_stackname_array[@]}; do
			attach_policy "${nested_own_stackname_array[$i]}" &
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

function build_upload(){
	local repository_name=$(echo "${CFN_PJ_PREFIX}-ECR" | tr '[:upper:]' '[:lower:]')
	local ecr_repository_enddpoint="${account_id}.dkr.ecr.ap-northeast-1.amazonaws.com"
	local ecr_repository_uri="${ecr_repository_enddpoint}/${repository_name}"

	local ecr_tag="test"

	docker build -t ${repository_name} .

	docker tag ${repository_name}:latest ${ecr_repository_uri}:${ecr_tag}

	aws ecr get-login-password --region ${REGION} ${profile_option} |
		docker login --username AWS --password-stdin ${ecr_repository_enddpoint}

	docker push ${ecr_repository_uri}:${ecr_tag}
}

function object_upload(){
	local own_stackname="${1}"

	local resources=$(
		aws cloudformation list-stack-resources \
			--stack-name ${own_stackname} \
			--query "StackResourceSummaries" \
			${profile_option} \
		| jq '.[] | {LogicalResourceId:.LogicalResourceId, PhysicalResourceId:.PhysicalResourceId, ResourceType:.ResourceType}' \
		| jq -s '.'
	)

	local bucket_resources=$(
		echo "${resources}" \
		| jq '.[] | select(.ResourceType == "AWS::S3::Bucket") | .PhysicalResourceId' \
		| jq -s '.'
	)

	local nested_stack_resources=$(
		echo "${resources}" \
		| jq '.[] | select(.ResourceType == "AWS::CloudFormation::Stack") | .PhysicalResourceId' \
		| jq -s '.'
	)

	local bucket_resource_len=$(echo $bucket_resources | jq length)
	local nested_stack_resourceLen=$(echo $nested_stack_resources | jq length)
	local bucket_name_array=()
	local nested_own_stackname_array=()


	if [ ${bucket_resource_len} -gt 0 ]; then
		for i in $( seq 0 $(($bucket_resource_len - 1)) ); do 
			bucket_name_array+=( $(echo $bucket_resources | jq -r ".[$i]") )
		done
	fi

	if [ ${nested_stack_resourceLen} -gt 0 ]; then
		for i in $( seq 0 $(($nested_stack_resourceLen - 1)) ); do 
			nested_own_stackname_array+=( $(
				echo $nested_stack_resources \
				| jq -r ".[$i]" \
				| sed -e "s/^arn:aws:cloudformation:[^:]*:[0-9]*:stack\/\([^\/]*\)\/.*$/\1/g" \
				) \
			)
		done

		local pids=()
		for i in ${!nested_own_stackname_array[@]}; do
			object_upload "${nested_own_stackname_array[$i]}" &
			pids[$!]=$!
		done
		wait ${pids[@]}
	fi

    filename="./Dockerfile"
	for i in ${!bucket_name_array[@]}; do
        aws s3 cp ${filename} s3://${bucket_name_array[$i]} ${profile_option}
	done
}


if [ -z "`aws s3 ls ${profile_option} | grep ${sam_bucket}`" ]; then
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


attach_policy "${CFN_STACK_NAME}"

object_upload "${CFN_STACK_NAME}"
object_upload "${CFN_STACK_NAME}" # update version

build_upload