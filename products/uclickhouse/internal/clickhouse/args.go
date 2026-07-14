package clickhouse

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func noArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	if len(args) == 1 && (args[0] == "true" || args[0] == "false") {
		return fmt.Errorf("unexpected argument %q for %s; boolean flags must use --flag=%s", args[0], cmd.CommandPath(), args[0])
	}
	return fmt.Errorf("unexpected argument(s) for %s: %s", cmd.CommandPath(), strings.Join(args, " "))
}

func noFlagLikeValues(cmd *cobra.Command, names ...string) error {
	for _, name := range names {
		flag := cmd.Flags().Lookup(name)
		if flag == nil || !flag.Changed {
			continue
		}
		value := flag.Value.String()
		if strings.HasPrefix(value, "-") {
			return fmt.Errorf("flag --%s requires a value; got %q, which looks like another flag", name, value)
		}
	}
	return nil
}

func requireFlagsWhenBool(cmd *cobra.Command, conditionName string, conditionValue bool, requiredNames ...string) error {
	value, err := cmd.Flags().GetBool(conditionName)
	if err != nil || value != conditionValue {
		return err
	}
	return requireNonEmptyFlags(cmd, fmt.Sprintf("--%s=%t", conditionName, conditionValue), requiredNames...)
}

func requireFlagsWhenString(cmd *cobra.Command, conditionName, conditionValue string, requiredNames ...string) error {
	value, err := cmd.Flags().GetString(conditionName)
	if err != nil || !strings.EqualFold(value, conditionValue) {
		return err
	}
	return requireNonEmptyFlags(cmd, fmt.Sprintf("--%s=%s", conditionName, conditionValue), requiredNames...)
}

func requireNonEmptyFlags(cmd *cobra.Command, condition string, names ...string) error {
	missing := []string{}
	for _, name := range names {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			continue
		}
		if isEmptyFlagValue(flag.Value.String()) {
			missing = append(missing, "--"+name)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("missing required flag(s) when %s: %s", condition, strings.Join(missing, ", "))
}

func isEmptyFlagValue(value string) bool {
	return value == "" || value == "[]" || value == "<nil>"
}
