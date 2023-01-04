package io

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/go-to-k/delstack/internal/resourcetype"
)

func DoInteractiveMode() ([]string, bool) {
	var checkboxes []string

	for {
		checkboxes = getCheckboxes()

		if len(checkboxes) == 0 {
			Logger.Warn().Msg("Select ResourceTypes!")
			ok := getYesNo("Do you want to finish?")
			if ok {
				Logger.Info().Msg("Finished...")
				return checkboxes, false
			}
			continue
		}

		ok := getYesNo("OK?")
		if ok {
			return checkboxes, true
		}
	}
}

func getCheckboxes() []string {
	label := "Select ResourceTypes you wish to delete even if DELETE_FAILED." +
		"\n" +
		"However, if resources of the selected ResourceTypes will not be DELETE_FAILED when the stack is deleted, the resources will be deleted even if you selected. " +
		"\n"
	opts := resourcetype.GetResourceTypes()
	res := []string{}

	prompt := &survey.MultiSelect{
		Message: label,
		Options: opts,
	}
	survey.AskOne(prompt, &res)

	return res
}

func getYesNo(label string) bool {
	choices := "Y/n"
	r := bufio.NewReader(os.Stdin)
	var s string

	for {
		fmt.Fprintf(os.Stderr, "%s (%s) ", label, choices)
		s, _ = r.ReadString('\n')
		fmt.Fprintln(os.Stderr)

		s = strings.TrimSpace(s)
		if s == "" {
			return true
		}
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		}
		if s == "n" || s == "no" {
			return false
		}
	}
}
