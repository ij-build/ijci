package db

type (
	PageMeta struct {
		Page     int
		PageSize int
	}

	PagedResultMeta struct {
		Total int
	}
)

func (m *PageMeta) Limit() int {
	return m.PageSize
}

func (m *PageMeta) Offset() int {
	return m.PageSize * (m.Page - 1)
}
