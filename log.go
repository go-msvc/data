package data

import (
	"github.com/jansemmelink/log"
)

//var log = logger.ForThisPackage()

//Debug ...
func Debug() {
	log.DebugOn()
	// logger.Top().WithWriter(file.NewFileWriter(os.Stderr))
	// e, _ := console.NewEncoder(console.Config{
	// 	Sep:  " ",
	// 	Term: "\n",
	// 	Columns: []console.ColumnConfig{
	// 		{Name: "timestamp"},
	// 		{Name: "level"},
	// 		{Name: "source"},
	// 		{Name: "text"},
	// 	}},
	// )
	// logger.Top().WithEncoder(e)
	// logger.Top().WithLevel(logger.InfoLevel)
	// log.WithLevel(logger.DebugLevel)
}
