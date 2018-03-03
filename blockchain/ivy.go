package blockchain

import (
	"strings"

	chainjson "github.com/bytom/encoding/json"
	"github.com/bytom/exp/ivy/compiler"
	"github.com/bytom/protocol/vm"
)

type (
	compileReq struct {
		Contract string                 `json:"contract"`
		Args     []compiler.ContractArg `json:"args"`
	}

	compileResp struct {
		Name    string             `json:"name"`
		Source  string             `json:"source"`
		Program chainjson.HexBytes `json:"program"`
		Params  []compiler.Param   `json:"params"`
		Value   string             `json:"value"`
		Clauses []clauseInfo       `json:"clause_info"`
		Opcodes string             `json:"opcodes"`
		Error   string             `json:"error"`
	}

	clauseInfo struct {
		Name      string               `json:"name"`
		Args      []compiler.Param     `json:"args"`
		Values    []compiler.ValueInfo `json:"value_info"`
		Mintimes  []string             `json:"mintimes"`
		Maxtimes  []string             `json:"maxtimes"`
		HashCalls []compiler.HashCall  `json:"hash_calls"`
	}
)

func compileIvy(req compileReq) (compileResp, error) {
	var resp compileResp
	compiled, err := compiler.Compile(strings.NewReader(req.Contract))
	if err != nil {
		resp.Error = err.Error()
	}

	// if source contract maybe contain import statement, multiple contract objects will be generated
	// after the compilation, and the last object is what we need.
	contract := compiled[len(compiled)-1]

	resp.Name = contract.Name
	resp.Source = req.Contract
	resp.Value = contract.Value
	resp.Opcodes = contract.Opcodes

	resp.Program = contract.Body
	if req.Args != nil {
		resp.Program, err = compiler.Instantiate(contract.Body, contract.Params, false, req.Args)
		if err != nil {
			resp.Error = err.Error()
		}

		resp.Opcodes, err = vm.Disassemble(resp.Program)
		if err != nil {
			return resp, err
		}
	}

	for _, param := range contract.Params {
		if param.InferredType != "" {
			param.Type = param.InferredType
			param.InferredType = ""
		}
		resp.Params = append(resp.Params, *param)
	}

	for _, clause := range contract.Clauses {
		info := clauseInfo{
			Name:      clause.Name,
			Args:      []compiler.Param{},
			Mintimes:  clause.MinTimes,
			Maxtimes:  clause.MaxTimes,
			HashCalls: clause.HashCalls,
		}
		if info.Mintimes == nil {
			info.Mintimes = []string{}
		}
		if info.Maxtimes == nil {
			info.Maxtimes = []string{}
		}

		for _, p := range clause.Params {
			info.Args = append(info.Args, compiler.Param{Name: p.Name, Type: p.Type})
		}

		for _, value := range clause.Values {
			info.Values = append(info.Values, value)
		}

		resp.Clauses = append(resp.Clauses, info)
	}

	return resp, nil
}
