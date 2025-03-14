# process mining app

in this version the program is a prototype for building a process graph.

to run the project you need to write `go run main.go`
then open the specified port on localhost

at this stage the program supports building a graph on a `csv file up to 1.5 GB`.

example of a csv file

```csv
{
    "SessionID",    - unique id of process
    "Timestamp",    - datetime of event
    "Description"   - name of event
}
```