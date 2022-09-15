package spry

import (
	"fmt"
	"reflect"
	"testing"
)

type Person struct {
	ids  Identifiers
	Name string
}

type NameChanged struct {
	Name string `json:"name"`
}

type ChangeName struct {
	Name string `json:"name"`
}

func (person Person) GetIdentifiers() Identifiers {
	return person.ids
}

func (person *Person) onNameChanged(event NameChanged) {
	person.Name = event.Name
}

func (person *Person) changeName(cmd ChangeName) ([]any, error) {
	return []any{
		NameChanged(cmd),
	}, nil
}

func (person Person) Apply(events []any) Person {
	for _, event := range events {
		switch ev := event.(type) {
		case NameChanged:
			person.onNameChanged(ev)
		}
	}
	return person
}

func (person Person) Handle(command any) ([]any, error) {
	switch cmd := command.(type) {
	case ChangeName:
		return person.changeName(cmd)
	}
	return nil,
		fmt.Errorf(
			"%T has no handler for %T",
			reflect.TypeOf(person),
			reflect.TypeOf(command),
		)
}

func TestGetRepositoryFor(t *testing.T) {
	repo := GetRepositoryFor[Person]()
	if repo.ActorName != "Person" {
		t.Error("actor name was not the expected type")
	}
}

func TestHandleCommandSuccessfully(t *testing.T) {
	repo := GetRepositoryFor[Person]()
	results := repo.Handle(ChangeName{Name: "Bob"})
	expected := NameChanged{Name: "Bob"}
	if len(results.Events) == 0 ||
		results.Events[0].(NameChanged) != expected {
		t.Error("event was not generated or did not match expected output")
	}
	if results.Original.Name != "" {
		t.Error("original actor instance was modified but should not have been")
	}
	if results.Modified.Name != "Bob" {
		t.Error("modified actor did not contain expected state")
	}
}
