package main
import (
	"fmt"
	"reflect"
	"github.com/caddyserver/certmagic"
)
func main() {
	var s certmagic.DNS01Solver
	t := reflect.TypeOf(s.DNSManager)
	fmt.Printf("DNSManager Fields:\n")
	for i := 0; i < t.NumField(); i++ {
		fmt.Printf("- %s\n", t.Field(i).Name)
	}
}
