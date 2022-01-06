package cmdx

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"
)

func SetRefCmd(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reference",
		Short: fmt.Sprint(root.Name(), " reference"),
		Long:  referenceLong(root),
		Run:   referenceHelpFn(),
	}
	cmd.SetHelpFunc(referenceHelpFn())
	return cmd
}

func referenceHelpFn() func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		md, err := printer.Markdown(cmd.Long)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Print(md)
	}
}

func referenceLong(cmd *cobra.Command) string {
	buf := bytes.NewBufferString(fmt.Sprintf("# %s reference\n\n", cmd.Name()))
	for _, c := range cmd.Commands() {
		if c.Hidden {
			continue
		}
		cmdRef(buf, c, 2)
	}
	return buf.String()
}

func cmdRef(w io.Writer, cmd *cobra.Command, depth int) {
	// Name + Description
	fmt.Fprintf(w, "%s `%s`\n\n", strings.Repeat("#", depth), cmd.UseLine())
	fmt.Fprintf(w, "%s\n\n", cmd.Short)

	if flagUsages := cmd.Flags().FlagUsages(); flagUsages != "" {
		fmt.Fprintf(w, "```\n%s````\n\n", dedent(flagUsages))
	}

	// Subcommands
	for _, c := range cmd.Commands() {
		if c.Hidden {
			continue
		}
		cmdRef(w, c, depth+1)
	}
}
