package data

//ISetter represents a name-value pair of data that can be modified
type ISetter interface {
	Set(name string, value interface{})
	Assign(assignments Assignments, from []IGetter) error
}
