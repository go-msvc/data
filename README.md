# Data #

Library functions to manipulate data.

# CSV #

Convert any type to CSV with:
```
values,err := data.CSV(myValue)
csvBuffer,err := bytes.NewBuffer(nil)
csv.NewWriter(csvBuffer).Write(values)
fmt.Printf("CSV: %s", string(csvBuffer.Bytes()))
```
