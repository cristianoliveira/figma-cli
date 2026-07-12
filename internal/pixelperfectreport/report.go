package pixelperfectreport

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/imagediff"
)

type Input struct {
	ReferencePath string
	ActualPath    string
	MaskPath      string
	OverlayPath   string
	Result        imagediff.ImageComparison
}

type view struct {
	Input
	ReferenceDataURL template.URL
	ActualDataURL    template.URL
	MaskDataURL      template.URL
	OverlayDataURL   template.URL
}

func Write(path string, input Input) error {
	html, err := Render(input)
	if err != nil {
		return err
	}
	return os.WriteFile(path, html, 0o600)
}

func Render(input Input) ([]byte, error) {
	view := view{Input: input}
	var err error
	view.ReferenceDataURL, err = imageDataURL(input.ReferencePath)
	if err != nil {
		return nil, err
	}
	view.ActualDataURL, err = imageDataURL(input.ActualPath)
	if err != nil {
		return nil, err
	}
	view.MaskDataURL, err = imageDataURL(input.MaskPath)
	if err != nil {
		return nil, err
	}
	if input.OverlayPath != "" {
		view.OverlayDataURL, err = imageDataURL(input.OverlayPath)
		if err != nil {
			return nil, err
		}
	}
	var output bytes.Buffer
	if err := reportTemplate.Execute(&output, view); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func imageDataURL(path string) (template.URL, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(data)), nil
}

var reportTemplate = template.Must(template.New("report").Parse(`<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>Pixel Perfect Report</title></head>
<body>
<h1>Pixel Perfect Report</h1>
<section><h2>Run configuration and provenance</h2>
<ul><li>Reference: {{.ReferencePath}}</li><li>Actual: {{.ActualPath}}</li><li>Mask: {{.MaskPath}}</li>{{if .OverlayPath}}<li>Overlay: {{.OverlayPath}}</li>{{end}}</ul>
</section>
<section><h2>Global metrics</h2>
<ul><li>Changed pixels: {{.Result.ChangedPixels}}</li><li>Changed ratio: {{.Result.ChangedRatio}}</li><li>RMSE: {{.Result.RMSE}}</li><li>Perceptual changed ratio: {{.Result.PerceptualChangedRatio}}</li></ul>
</section>
<section><h2>Artifacts</h2>
<h3>Reference</h3><img alt="Reference" src="{{.ReferenceDataURL}}">
<h3>Actual</h3><img alt="Actual" src="{{.ActualDataURL}}">
<h3>Mask</h3><img alt="Mask" src="{{.MaskDataURL}}">
{{if .OverlayDataURL}}<h3>Overlay</h3><img alt="Overlay" src="{{.OverlayDataURL}}">{{end}}
</section>
<section><h2>Ranked deterministic regions</h2>
{{if .Result.Regions}}<ol>{{range .Result.Regions}}<li>Bounds: {{.Bounds.X}},{{.Bounds.Y}},{{.Bounds.Width}},{{.Bounds.Height}} Changed: {{.ChangedPixels}} Classification: {{.Classification}}</li>{{end}}</ol>{{else}}<p>No changed regions.</p>{{end}}
</section>
{{if .Result.SuggestedOffset}}<section><h2>Suggested offset</h2><p>x={{.Result.SuggestedOffset.X}} y={{.Result.SuggestedOffset.Y}} rmse={{.Result.SuggestedOffset.RMSE}}</p></section>{{end}}
</body></html>
`))
