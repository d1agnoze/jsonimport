package server

type completionList struct {
	/*
		This list is not complete. Further typing should result in recomputing
		this list.

		Recomputed lists have all their items replaced (not appended) in the
		incomplete completion sessions.
	*/
	IsIncomplete bool `json:"isIncomplete"`

	//The completion items.
	Items []completionItem `json:"items"`
}

type completionItem struct {
	Label         string                      `json:"label,omitempty"`
	LabelDetail   *completionItemLabelDetails `json:"labelDetails,omitempty"`
	Kind          completionKind              `json:"kind,omitzero"`
	Detail        *string                     `json:"detail"`
	Documentation *string                     `json:"documentation"`
	Preselect     *bool                       `json:"preselect"`
	InsertText    *string                     `json:"insertText"`
}

type completionItemLabelDetails struct {
	Detail       string `json:"detail,omitempty"`
	Descriptionn string `json:"description,omitempty"`
}

type completionKind int

func (c completionKind) IsZero() bool {
	return c == 0
}

const (
	completionKindOmit completionKind = iota
	completionKindText
	completionKindMethod
	completionKindFunction
	completionKindConstructor
	completionKindField
	completionKindVariable
	completionKindClass
	completionKindInterface
	completionKindModule
	completionKindProperty
	completionKindUnit
	completionKindValue
	completionKindEnum
	completionKindKeyword
	completionKindSnippet
	completionKindColor
	completionKindFile
	completionKindReference
	completionKindFolder
	completionKindEnumMember
	completionKindConstant
	completionKindStruct
	completionKindEvent
	completionKindOperator
	completionKindTypeParameter
)
