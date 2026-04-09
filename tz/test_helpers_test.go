package tz

func makeTestContext() *RenderContext {
	return &RenderContext{app: &App{componentStates: make(map[string]any)}}
}
