# feature/1-init-project

in this version the program is a prototype for building a process graph.

to run the project you need to write `go run main.go`
then open the specified port on localhost

at this stage the program supports building a graph on a `csv file up to 1 GB`.

example of a csv file

```csv
{
    "SessionID", 
    "Timestamp", 
    "Description"
}
```