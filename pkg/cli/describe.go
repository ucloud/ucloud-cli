package cli

// DescribeRow is a single attribute/content row for rendering single-resource
// detail views (e.g. `... describe`). When passed to Context.PrintList in table
// mode, the field names "Attribute" and "Content" become the column headers.
//
// It is defined standalone (not aliased to base.DescribeTableRow) on purpose, to
// keep pkg/cli free of a dependency on the base package.
type DescribeRow struct {
	Attribute string
	Content   string
}
