package html

import (
	"bytes"
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
	tpls "github.com/go-echarts/go-echarts/v2/templates"
	"github.com/mqufflc/whodidthechores/internal/repository"
)

var chartTpl = `
{{- define "chart" }}
{{- range .JSAssets.Values }}
    <script src="{{ . }}"></script>
{{- end }}
{{- template "base" . }}
<script type="text/javascript">
	window.addEventListener('resize', function() {
		goecharts_{{ .ChartID | safeJS }}.resize();
	});
</script>
<style>
    .container {padding-bottom: 30px;display: flex;justify-content: center;align-items: center; height: 100%; width: 100%;}
    .item {margin: auto;}
</style>
{{ end }}
`

var baseTpl = `
{{- define "base_element" -}}
<div class="container">
    <div class="item" id="{{ .ChartID }}" style="min-width: 100%; min-height: 100%;"></div>
</div>
{{- end -}}

{{- define "base_script" -}}
<script type="text/javascript">
    "use strict";

	let goecharts_{{ .ChartID | safeJS }} = echarts.init(document.getElementById('{{ .ChartID | safeJS }}'), "{{ .Theme }}", { renderer: "{{  .Initialization.Renderer }}" });
	let option_{{ .ChartID | safeJS }} = {{ template "base_option" . }}
	goecharts_{{ .ChartID | safeJS }}.setOption(option_{{ .ChartID | safeJS }});

	document.addEventListener("DOMContentLoaded", () => {
		setTimeout(function () {
			goecharts_{{ .ChartID | safeJS }}.resize()
		}, 100)
	});

  {{- range  $listener := .EventListeners }}
    {{if .Query  }}
    goecharts_{{ $.ChartID | safeJS }}.on({{ $listener.EventName }}, {{ $listener.Query | safeJS }}, {{ injectInstance $listener.Handler "%MY_ECHARTS%"  $.ChartID | safeJS }});
    {{ else }}
    goecharts_{{ $.ChartID | safeJS }}.on({{ $listener.EventName }}, {{ injectInstance $listener.Handler "%MY_ECHARTS%"  $.ChartID | safeJS }})
    {{ end }}
  {{- end }}

    {{- range .JSFunctions.Fns }}
    {{ injectInstance . "%MY_ECHARTS%"  $.ChartID  | safeJS }}
    {{- end }}
</script>
{{- end -}}

{{- define "base_option" }}
    {{- .JSONNotEscaped | safeJS }}
{{- end }};

{{- define "base" }}
    {{- template "base_element" . }}
    {{- template "base_script" . }}
{{- end }}
`

type snippetRenderer struct {
	render.BaseRender
	c      interface{}
	before []func()
}

func NewSnippetRenderer(c interface{}, before ...func()) render.Renderer {
	return &snippetRenderer{c: c, before: before}
}

func (r *snippetRenderer) Render(w io.Writer) error {
	for _, fn := range r.before {
		fn()
	}

	contents := []string{tpls.HeaderTpl, baseTpl, chartTpl}
	tpl := render.MustTemplate("chart", contents)

	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "chart", r.c); err != nil {
		return err
	}

	_, err := w.Write(buf.Bytes())
	return err
}

func generateBarItems(chore string, report repository.Report) []opts.BarData {
	items := make([]opts.BarData, 0)
	for _, user := range report.Users {
		items = append(items, opts.BarData{Value: report.Report[chore][user]})
	}
	return items
}

func CreateBarChart(report repository.Report) *charts.Bar {
	bar := charts.NewBar()
	bar.Renderer = NewSnippetRenderer(bar, bar.Validate)
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(
			opts.Initialization{AssetsHost: "/static/"},
		),
		charts.WithLegendOpts(
			opts.Legend{Type: "scroll", Show: opts.Bool(true), Bottom: "20"},
		),
	)
	bar.SetXAxis(report.Users)
	for _, chore := range report.Chores {
		bar.AddSeries(chore, generateBarItems(chore, report))
	}
	bar.SetSeriesOptions(charts.WithBarChartOpts(opts.BarChart{
		Stack: "stackA",
	}))
	return bar
}

// The charts all have a `Render(w io.Writer) error` method on them.
// That method is very similar to templ's Render method.
type Renderable interface {
	Render(w io.Writer) error
}

// So lets adapt it.
func ConvertChartToTemplComponent(chart Renderable) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return chart.Render(w)
	})
}
