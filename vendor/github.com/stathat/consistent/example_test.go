// Copyright (C) 2012 Numerotron Inc.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.

package consistent_test

import (
	"consistent"
	"fmt"
	"log"
)

func ExampleNew() {
	c := consistent.New()
	c.Add("cacheA")
	c.Add("cacheB")
	c.Add("cacheC")
	users := []string{"user_mcnulty", "user_bunk", "user_omar", "user_bunny", "user_stringer"}
	for _, u := range users {
		server, err := c.Get(u)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s => %s\n", u, server)
	}
	// Output:
	// user_mcnulty => cacheA
	// user_bunk => cacheC
	// user_omar => cacheC
	// user_bunny => cacheA
	// user_stringer => cacheB
}

func ExampleAdd() {
	c := consistent.New()
	c.Add("cacheA")
	c.Add("cacheB")
	c.Add("cacheC")
	users := []string{"user_mcnulty", "user_bunk", "user_omar", "user_bunny", "user_stringer"}
	fmt.Println("initial state [A, B, C]")
	for _, u := range users {
		server, err := c.Get(u)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s => %s\n", u, server)
	}
	c.Add("cacheD")
	c.Add("cacheE")
	fmt.Println("\nwith cacheD, cacheE [A, B, C, D, E]")
	for _, u := range users {
		server, err := c.Get(u)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s => %s\n", u, server)
	}
	// Output:
	// initial state [A, B, C]
	// user_mcnulty => cacheA
	// user_bunk => cacheC
	// user_omar => cacheC
	// user_bunny => cacheA
	// user_stringer => cacheB
	//
	// with cacheD, cacheE [A, B, C, D, E]
	// user_mcnulty => cacheE
	// user_bunk => cacheD
	// user_omar => cacheD
	// user_bunny => cacheA
	// user_stringer => cacheB
}

func ExampleRemove() {
	c := consistent.New()
	c.Add("cacheA")
	c.Add("cacheB")
	c.Add("cacheC")
	users := []string{"user_mcnulty", "user_bunk", "user_omar", "user_bunny", "user_stringer"}
	fmt.Println("initial state [A, B, C]")
	for _, u := range users {
		server, err := c.Get(u)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s => %s\n", u, server)
	}
	c.Remove("cacheC")
	fmt.Println("\ncacheC removed [A, B]")
	for _, u := range users {
		server, err := c.Get(u)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s => %s\n", u, server)
	}
	// Output:
	// initial state [A, B, C]
	// user_mcnulty => cacheA
	// user_bunk => cacheC
	// user_omar => cacheC
	// user_bunny => cacheA
	// user_stringer => cacheB
	//
	// cacheC removed [A, B]
	// user_mcnulty => cacheA
	// user_bunk => cacheA
	// user_omar => cacheA
	// user_bunny => cacheA
	// user_stringer => cacheB
}
