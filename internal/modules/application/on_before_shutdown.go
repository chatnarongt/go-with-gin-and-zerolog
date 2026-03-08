package application

func (m *Module) OnBeforeShutdown(f ...func()) {
	m.onBeforeShutdowns = append(m.onBeforeShutdowns, f...)
}
