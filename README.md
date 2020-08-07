# deviceNTask

Basic usage is just to import the library as the following example:

```go
import (
	"log"

	"github.com/PacoDw/deviceNTask/dnt"
)

func main() {
	if err := dnt.CreateOptimalConfigurationFile("challenge.in"); err != nil {
		log.Fatal(err)
	}
}
```

If you want to run the current example, clone the repository and once you are in the folder project open the terminal and write the following command to run it:
```bash
$ go run main.go
``` 
And the `challenge.out` file will generate will the results

Note: you need the file `challenge.in` in order to work fine
