package dnt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/spf13/cast"
)

// Resource contains the ID and Consumption
type Resource struct {
	ID          int
	Consumption int
}

// OptimalConfig represents the final results describing the best configuration
type OptimalConfig struct {
	Backgroud  *Resource
	Foreground *Resource
	Total      int
}

// Config saves all the configuration provided by the challenge.in so this struct
// will allows to manipulate that information to ceonvert them to the final result
type Config struct {
	ID            int
	Capacity      int
	Foreground    []Resource
	Backgroud     []Resource
	OptimalConfig []OptimalConfig

	metadata []byte
}

// toByte will convert the OptimalConfig struct to a []byte to be write in challenge.out
func (c *Config) toByte() []byte {
	res := []string{}

	for i := range c.OptimalConfig {
		if !reflect.DeepEqual(c.OptimalConfig[i], OptimalConfig{}) {
			res = append(res, fmt.Sprintf(`(%d, %d)`, c.OptimalConfig[i].Foreground.ID, c.OptimalConfig[i].Backgroud.ID))
		}
	}

	return []byte(strings.Join(res, ", ") + "\n")
}

// newConfig creates a new configuration returning a &Config{} struct
func newConfig(id int, metadata []byte) *Config {
	return &Config{
		ID:       id,
		metadata: metadata,
	}
}

// createOptimalConfig analize which configuration is optimal or not
func (c *Config) createOptimalConfig() *Config {
	ocs := []OptimalConfig{}

	if len(c.Foreground) >= len(c.Backgroud) {

		for i := range c.Foreground {
			fr := c.Foreground[i]

			for j := range c.Backgroud {
				br := c.Backgroud[j]

				total := fr.Consumption + br.Consumption

				if total <= c.Capacity {

					oc := OptimalConfig{
						Foreground: &fr,
						Backgroud:  &br,
						Total:      total,
					}

					if len(ocs) > 0 {
						if oc.Total > ocs[0].Total {
							ocs = append([]OptimalConfig{oc}, OptimalConfig{})
						} else if oc.Total == ocs[0].Total {
							ocs = append([]OptimalConfig{oc}, ocs...)
						}
					} else {
						ocs = append(ocs, oc)
					}
				}
			}
		}
	}

	c.OptimalConfig = ocs

	return c
}

// createConfigs can acept the file in order to separate each configuration and storage in
// each Config struct, when one of them is ready it will passes bye the chanel
func createConfigs(r io.Reader) <-chan *Config {
	var (
		count, index = 0, 0
		scanner      = bufio.NewScanner(r)
		ConfigCh     = make(chan *Config)
		metadata     = []byte{}
		wg           sync.WaitGroup
	)

	// we need to read each line of the file to identify each config
	scanner.Split(bufio.ScanLines)

	go func() {
		for scanner.Scan() {
			wg.Add(1)
			count++

			// the scanner will return the each line but it removes the new line so
			// for this reason it's putted again
			metadata = append(metadata, append(scanner.Bytes(), "\n"...)...)

			// for each line will determinate that's is a new configuration
			if count%3 == 0 {
				index++
				ConfigCh <- newConfig(index, metadata)
				metadata = []byte{}
				count = 0
			}

			wg.Done()
		}
		wg.Wait()
		close(ConfigCh)
	}()

	return ConfigCh
}

// newResources converts the string of the file to a []Resource{} struct
func newResources(b []byte) []Resource {
	reg := regexp.MustCompile("[0-9]+")
	numbers := cast.ToIntSlice(reg.FindAllString(string(b), -1))

	res := make([]Resource, len(numbers)/2)

	for i := 1; i < len(res)+1; i++ {
		pars := numbers[i*2-2 : i*2]
		res[i-1].ID = pars[0]
		res[i-1].Consumption = pars[1]
	}

	return res
}

// setFields will set each field of the Resource struct: Capacity, Foreground and Backgroud
func (c *Config) setFields() *Config {
	scanner := bufio.NewScanner(bytes.NewReader(c.metadata))
	scanner.Split(bufio.ScanLines)
	count := 0

	for scanner.Scan() {
		switch count {
		case 0:
			c.Capacity = cast.ToInt(scanner.Text())
		case 1:
			c.Foreground = newResources(scanner.Bytes())
		case 2:
			c.Backgroud = newResources(scanner.Bytes())
		}

		count++
	}

	return c
}

// getOptimalConfiguration this is a private method that returns a chanel with each final Optimal
// config for each configuration
func getOptimalConfiguration(filename string) (*chan []byte, error) {
	// read the file
	file, err := os.Open("challenge.in")
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	// call the create method, it retrieves a chanel with each configuration
	//  note: this method convert the file to each config
	ConfigChan := createConfigs(file)
	// configs will be returned with the finals results
	configs := make(chan []byte)

	go func() {

		for {
			d, ok := <-ConfigChan
			if !ok {
				break
			}
			wg.Add(1)

			// for each config we will convert each one to the final result as a []binary
			configs <- d.setFields().createOptimalConfig().toByte()
			wg.Done()
		}

		wg.Wait()

		close(configs)
	}()

	return &configs, nil
}

// CreateOptimalConfigurationFile will return the final result and create the file with the most
// optimal configurations
func CreateOptimalConfigurationFile(filename string) error {
	configs, err := getOptimalConfiguration(filename)
	if err != nil {
		return err
	}

	f, err := os.OpenFile("challenge.out", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for {
		w, ok := <-*configs
		if !ok {
			break
		}
		_, err = f.Write(w)
		if err != nil {
			return err
		}
	}

	return nil
}
