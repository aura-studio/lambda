package tree_test

import (
	"fmt"
	"testing"

	"github.com/aura-studio/lambda/boost/magic"
	"github.com/aura-studio/lambda/boost/tree"
	"github.com/aura-studio/lambda/core/encoding"
)

type Zoo struct {
	Name    string
	Keepers []Keeper
	Animals []Animal
}

type Keeper struct {
	Name string
	Age  int
}

type Animal struct {
	SpeciesName string
	Count       int
}

func TestMapTree(t *testing.T) {
	var zoo = Zoo{
		Name: "Beasts & Monsters",
		Keepers: []Keeper{
			{
				Name: "Alice",
				Age:  20,
			},
			{
				Name: "Tony",
				Age:  50,
			},
		},
		Animals: []Animal{
			{
				SpeciesName: "Elephant",
				Count:       3,
			},
			{
				SpeciesName: "Money",
				Count:       20,
			},
		},
	}

	json := encoding.NewJSON()
	zooJSON, err := json.Marshal(zoo)
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]interface{}
	err = json.Unmarshal(zooJSON, &data)
	if err != nil {
		t.Fatal(err)
	}
	tree1 := tree.NewTree()
	tree1.SetData(data)

	v := tree1.Get(magic.UnixChain("animals.0.species_name"))
	if s, ok := v.(string); !ok {
		t.Fatalf("expected a string, but got %v", v)
	} else if s != "Elephant" {
		t.Fatalf("expected a string of `Elephant`, but got %s", s)
	}

	tree1.Set(magic.UnixChain("name"), "Sea World")
	tree1.Set(magic.UnixChain("animals.0.species_name"), "Dolphin")
	tree1.Set(magic.UnixChain("keepers.3"), map[string]interface{}{"Name": "Mike", "Age": 29})
	tree1.Set(magic.UnixChain("animals.4.1.3"), "a mistake")

	if tree1.Get(magic.UnixChain("name")) != "Sea World" {
		t.Fatalf("expected zoo name `Sea World`, but got %s", tree1.Get(magic.UnixChain("sea")))
	}
	if tree1.Get(magic.UnixChain("animals.0.species_name")) != "Dolphin" {
		t.Fatalf("expected species name `Dolphin`, but got %s", tree1.Get(magic.UnixChain("animals.0.species_name")))
	}
	if tree1.Get(magic.UnixChain("keepers.3.name")) != "Mike" {
		t.Fatalf("expected keeper name `Mike`, but got %s", tree1.Get(magic.UnixChain("keepers.3.name")))
	}

	tree1.Remove(magic.UnixChain("animals.4.1.3"))
	tree1.Remove(magic.UnixChain("Keepers.1.Name"))
	tree1.Remove(magic.UnixChain("Keepers.1.Age"))
	zooJSON, err = json.Marshal(tree1.Data())
	if err != nil {
		t.Fatal(err)
	}
	zoo = Zoo{}
	err = json.Unmarshal(zooJSON, &zoo)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", zoo)
	if zoo.Name != "Sea World" {
		t.Fatalf("expected species name `Sea World`, but got %s", zoo.Name)
	}
	if zoo.Animals[0].SpeciesName != "Dolphin" {
		t.Fatalf("expected species name `Dolphin`, but got %s", zoo.Animals[0].SpeciesName)
	}
	if zoo.Keepers[3].Name != "Mike" {
		t.Fatalf("expected keeper name `Mike`, but got %s", zoo.Keepers[3].Name)
	}

	mapTree2 := tree1.Dulplicate()

	mapTree2.Remove(magic.UnixChain("Keepers.3.Name"))
	mapTree2.Remove(magic.UnixChain("Keepers.3.Age"))
	mapTree2.Set(magic.UnixChain("Name"), "Universe World")
	mapTree2.Set(magic.UnixChain("Keepers.0.Age"), 100)
	zooJSON, err = json.Marshal(mapTree2.Data())
	if err != nil {
		t.Fatal(err)
	}
	zoo = Zoo{}
	err = json.Unmarshal(zooJSON, &zoo)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", zoo)
	if len(zoo.Keepers) != 1 {
		t.Fatalf("expected 1 keeper, but got %d", len(zoo.Keepers))
	}

	tree1.Merge(mapTree2)
	zooJSON, err = json.Marshal(tree1.Data())
	if err != nil {
		t.Fatal(err)
	}
	zoo = Zoo{}
	err = json.Unmarshal(zooJSON, &zoo)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", zoo)
	if zoo.Name != "Universe World" {
		t.Fatalf("expected name `Universe World`, but got %s", zoo.Name)
	}

	hash, err := encoding.NewHash().Marshal(tree1)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(hash))

	tree3 := tree.NewTree()
	if err := encoding.NewHash().Unmarshal(hash, tree3); err != nil {
		t.Fatal(err)
	}

	// DeepEqual doesnot match
	if fmt.Sprintf("%v", tree3) != fmt.Sprintf("%v", tree1) {
		t.Fatalf("expected `mapTree3` equals to `tree`, mapTree3: %+v, tree: %+v", tree3, tree1)
	}
}
