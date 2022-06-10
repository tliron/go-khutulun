package agent

import (
	"github.com/tliron/kutil/ard"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Agent) Instantiate(clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) bool {
	// TODO apply redundancy policies

	for _, vertex := range clout.Vertexes {
		if types, ok := ard.NewNode(vertex.Properties).Get("types").StringMap(); ok {
			if _, ok := types["cloud.puccini.khutulun::Instantiated"]; ok {
				name, _ := ard.NewNode(vertex.Properties).Get("name").String()
				attributes, _ := ard.NewNode(vertex.Properties).Get("attributes").StringMap()
				attributes["instances"] = ard.StringMap{
					"$information": ard.StringMap{
						"entry": ard.StringMap{
							"type": ard.StringMap{"name": "string"},
						},
						"type": ard.StringMap{"name": "list"},
					},
					"$list": ard.List{
						ard.StringMap{
							"$map": ard.List{
								ard.Map{
									"$information": ard.StringMap{
										"type": ard.StringMap{
											"name": "string",
										},
									},
									"$key": ard.StringMap{
										"$value": "name",
									},
									"$value": name + "-0",
								},
							},
						},
						ard.StringMap{
							"$map": ard.List{
								ard.Map{
									"$information": ard.StringMap{
										"type": ard.StringMap{
											"name": "string",
										},
									},
									"$key": ard.StringMap{
										"$value": "host",
									},
									"$value": "here",
								},
							},
						},
					},
				}
			}
		}
	}

	return true // changed
}

/*
   $information:
     entry:
       name: cloud.puccini.khutulun::Instance
     type:
       name: list
   $list:
     - $map:
         - $information:
             type:
               name: string
           $key:
             $value: name
           $value: hello
         - $information:
             type:
               name: string
           $key:
             $value: host
           $value: here
*/
