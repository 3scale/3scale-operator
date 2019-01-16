package component

type MetaComponent struct {
	components []Component
}

type ImportableComponent interface {
	Import()
}

type ExportableComponent interface {
	Export()
}

type TemplatizableComponent interface {
	Templatize()
}

type PreAssemblableComponent interface {
	PreAssemble()
}

type PostAssemblableComponent interface {
	PostAssemble()
}

type ParametrizableComponent interface {
	Parametrize()
}

type DoubleBraceExpandableComponent interface {
	DoubleBraceExpand()
}

type GenericComponent interface {
	ImportableComponent
	ExportableComponent
	TemplatizableComponent
	PreAssemblableComponent
	PostAssemblableComponent
	ParametrizableComponent
	DoubleBraceExpandableComponent
}
