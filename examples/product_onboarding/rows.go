package onboarding

// instanceRow is the output struct for `example list`. When passed to
// ctx.PrintList in table mode, the exported field NAMES become the column
// headers, in declaration order. Keep the set small and human-meaningful: this
// is the at-a-glance view, not the full resource dump (that is `describe`).
type instanceRow struct {
	ResourceID string
	Name       string
	Zone       string
	Mode       string
	Spec       string
	Status     string
}
