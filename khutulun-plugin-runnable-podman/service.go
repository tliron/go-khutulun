package main

import (
	"strings"
)

//
// ServiceIdentifier
//

type ServiceIdentifier struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func (self *ServiceIdentifier) Equals(identifier *ServiceIdentifier) bool {
	if self == identifier {
		return true
	} else {
		return (self.Namespace == identifier.Namespace) && (self.Name == identifier.Name)
	}
}

// fmt.Stringer interface
func (self *ServiceIdentifier) String() string {
	return self.Namespace + "," + self.Name
}

//
// ServiceIdentifiers
//

type ServiceIdentifiers struct {
	List []*ServiceIdentifier
}

func NewServiceIdentifiers() *ServiceIdentifiers {
	return new(ServiceIdentifiers)
}

func (self *ServiceIdentifiers) Has(identifier *ServiceIdentifier) bool {
	for _, identifier_ := range self.List {
		if identifier_.Equals(identifier) {
			return true
		}
	}
	return false
}

func (self *ServiceIdentifiers) Add(identifiers ...*ServiceIdentifier) bool {
	var added bool
	for _, identifier := range identifiers {
		if !self.Has(identifier) {
			self.List = append(self.List, identifier)
			added = true
		}
	}
	return added
}

func (self *ServiceIdentifiers) Merge(identifiers *ServiceIdentifiers) bool {
	if identifiers != nil {
		return self.Add(identifiers.List...)
	} else {
		return false
	}
}

// fmt.Stringer interface
func (self *ServiceIdentifiers) String() string {
	var builder strings.Builder
	last := len(self.List) - 1
	for index, identifier := range self.List {
		builder.WriteString(identifier.String())
		if index != last {
			builder.WriteRune(';')
		}
	}
	return builder.String()
}
