package types

// Tables
type TableAlreadyExistsError struct {
	TableName string
}

func (e *TableAlreadyExistsError) Error() string {
	return "table '" + e.TableName + "' already exists"
}

type TableNotFoundError struct {
	TableName string
}

func (e *TableNotFoundError) Error() string {
	return "table '" + e.TableName + "' not found"
}

// Fields
type ReservedFieldError struct {
	FieldName string
}

func (e *ReservedFieldError) Error() string {
	return "cannot use reserved field: " + e.FieldName + ", we manage these fields automatically for you: id, created_at, updated_at"
}

type WrongFieldTypeError struct {
	FieldName    string
	ExpectedType string
	ActualType   string
}

func (e *WrongFieldTypeError) Error() string {
	return "wrong type for field '" + e.FieldName + "': expected " + e.ExpectedType + ", got " + e.ActualType
}

// Records
type RecordNotFoundError struct {
	TableName string
	RecordID  string
}

func (e *RecordNotFoundError) Error() string {
	return "table '" + e.TableName + "' has no record with ID '" + e.RecordID + "'"
}

// Values
type UniqueDuplicateValueError struct {
	FieldName string
	Value     string
}

func (e *UniqueDuplicateValueError) Error() string {
	return "duplicate value: '" + e.Value + "' for unique field: " + e.FieldName
}
