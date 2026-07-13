package image

// ImageRow 表格行 — mirrors uhost's ImageRow for ulhost image list display.
type ImageRow struct {
	ImageName         string
	ImageID           string
	ImageType         string
	BasicImage        string
	ExtensibleFeature string
	CreationTime      string
	State             string
}
