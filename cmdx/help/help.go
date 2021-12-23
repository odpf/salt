package help

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// RootUsageFunc provides a custom usage function for the root command.
func RootUsageFunc(command *cobra.Command) error {
	command.Printf("Usage:  %s", command.UseLine())

	subcommands := command.Commands()
	if len(subcommands) > 0 {
		command.Print("\n\nAvailable commands:\n")
		for _, c := range subcommands {
			if c.Hidden {
				continue
			}
			command.Printf("  %s\n", c.Name())
		}
		return nil
	}

	flagUsages := command.LocalFlags().FlagUsages()
	if flagUsages != "" {
		command.Println("\n\nFlags:")
		command.Print(indent(dedent(flagUsages), "  "))
	}
	return nil
}

// RootFlagErrorFunc provides a custom flag error function for the root command.
func RootFlagErrorFunc(cmd *cobra.Command, err error) error {
	if err == pflag.ErrHelp {
		return err
	}
	return err
}

// RootHelpFunc provides a custom help function for the root command.
func RootHelpFunc(command *cobra.Command, args []string) {
	if isRootCmd(command.Parent()) && len(args) >= 2 && args[1] != "--help" && args[1] != "-h" {
		nestedSuggestFunc(command, args[1])
		return
	}

	coreCommands := []string{}
	otherCommands := map[string][]string{}
	additionalCommands := []string{}

	for _, c := range command.Commands() {
		if c.Short == "" {
			continue
		}
		if c.Hidden {
			continue
		}

		s := rpad(c.Name(), c.NamePadding()) + c.Short
		if _, ok := c.Annotations["group:core"]; ok {
			coreCommands = append(coreCommands, s)
		} else if g, ok := c.Annotations["group:other"]; ok {
			otherCommands[g] = append(otherCommands[g], s)
		} else {
			additionalCommands = append(additionalCommands, s)
		}
	}

	// If there are no core and other commands, assume everything is a core command
	if len(coreCommands) == 0 && len(otherCommands) == 0 {
		coreCommands = additionalCommands
		additionalCommands = []string{}
	}

	type helpEntry struct {
		Title string
		Body  string
	}

	text := command.Long

	if text == "" {
		text = command.Short
	}

	helpEntries := []helpEntry{}
	if text != "" {
		helpEntries = append(helpEntries, helpEntry{"", text})
	}

	helpEntries = append(helpEntries, helpEntry{"USAGE", command.UseLine()})

	if len(coreCommands) > 0 {
		helpEntries = append(helpEntries, helpEntry{"CORE COMMANDS", strings.Join(coreCommands, "\n")})
	}

	for name, cmds := range otherCommands {
		if len(cmds) > 0 {
			helpEntries = append(helpEntries, helpEntry{strings.ToUpper(name) + " COMMANDS", strings.Join(cmds, "\n")})
		}
	}

	if len(additionalCommands) > 0 {
		helpEntries = append(helpEntries, helpEntry{"ADDITIONAL COMMANDS", strings.Join(additionalCommands, "\n")})
	}

	flagUsages := command.LocalFlags().FlagUsages()
	if flagUsages != "" {
		helpEntries = append(helpEntries, helpEntry{"FLAGS", dedent(flagUsages)})
	}

	inheritedFlagUsages := command.InheritedFlags().FlagUsages()
	if inheritedFlagUsages != "" {
		helpEntries = append(helpEntries, helpEntry{"INHERITED FLAGS", dedent(inheritedFlagUsages)})
	}

	if _, ok := command.Annotations["help:arguments"]; ok {
		helpEntries = append(helpEntries, helpEntry{"ARGUMENTS", command.Annotations["help:arguments"]})
	}

	if command.Example != "" {
		helpEntries = append(helpEntries, helpEntry{"EXAMPLES", command.Example})
	}

	if _, ok := command.Annotations["help:learn"]; ok {
		helpEntries = append(helpEntries, helpEntry{"LEARN MORE", command.Annotations["help:learn"]})
	}

	if _, ok := command.Annotations["help:feedback"]; ok {
		helpEntries = append(helpEntries, helpEntry{"FEEDBACK", command.Annotations["help:feedback"]})
	}

	out := command.OutOrStdout()
	for _, e := range helpEntries {
		if e.Title != "" {
			// If there is a title, add indentation to each line in the body
			fmt.Fprintln(out, e.Title)
			fmt.Fprintln(out, indent(strings.Trim(e.Body, "\r\n"), "  "))
		} else {
			// If there is no title print the body as is
			fmt.Println(e.Body)
		}
		fmt.Fprintln(out)
	}
}

// Display helpful error message in case subcommand name was mistyped.
// This matches Cobra's behavior for root command, which Cobra
// confusingly doesn't apply to nested commands.
func nestedSuggestFunc(command *cobra.Command, arg string) {
	command.Printf("unknown command %q for %q\n", arg, command.CommandPath())

	var candidates []string
	if arg == "help" {
		candidates = []string{"--help"}
	} else {
		if command.SuggestionsMinimumDistance <= 0 {
			command.SuggestionsMinimumDistance = 2
		}
		candidates = command.SuggestionsFor(arg)
	}

	if len(candidates) > 0 {
		command.Print("\nDid you mean this?\n")
		for _, c := range candidates {
			command.Printf("\t%s\n", c)
		}
	}

	command.Print("\n")
	_ = RootUsageFunc(command)
}

func isRootCmd(command *cobra.Command) bool {
	return command != nil && !command.HasParent()
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds ", padding)
	return fmt.Sprintf(template, s)
}

func dedent(s string) string {
	lines := strings.Split(s, "\n")
	minIndent := -1

	for _, l := range lines {
		if len(l) == 0 {
			continue
		}

		indent := len(l) - len(strings.TrimLeft(l, " "))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return s
	}

	var buf bytes.Buffer
	for _, l := range lines {
		fmt.Fprintln(&buf, strings.TrimPrefix(l, strings.Repeat(" ", minIndent)))
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

var lineRE = regexp.MustCompile(`(?m)^`)

func indent(s, indent string) string {
	if len(strings.TrimSpace(s)) == 0 {
		return s
	}
	return lineRE.ReplaceAllLiteralString(s, indent)
}
