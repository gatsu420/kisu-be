package commontype

type ToolType string

const (
	TableToolType ToolType = "table"
	QueryToolType ToolType = "query"
)

func (t ToolType) ValidateToolType() bool {
	switch t {
	case TableToolType, QueryToolType:
		return true
	}

	return false
}
