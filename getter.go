package data

//IGetter represents a name-value pair of data that can be retrieved only
type IGetter interface {
	Get(name string) (interface{}, error) //nil if not found, error if failed to get
}
