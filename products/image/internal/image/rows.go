package image

// ImageRow 表格行 — byte-identical to cmd/image.go's ImageRow.
type ImageRow struct {
	ImageName         string
	ImageID           string
	ImageType         string
	BasicImage        string
	ExtensibleFeature string
	CreationTime      string
	State             string
}
