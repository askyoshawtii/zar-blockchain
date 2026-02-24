package main
import (
	"fmt"
	"reflect"
	"github.com/caddyserver/certmagic"
)
func main() {
	t := reflect.TypeOf(certmagic.DNS01Solver{})
	fmt.Printf("DNS01Solver Fields:\n")
	for i := 0; i < t.NumField(); i++ {
		fmt.Printf("- %s\n", t.Field(i).Name)
	}
}
