package jobdist

import (
	"fmt"
	"net/http"
	"testing"
)

var (
	test_template interface{} = map[string]interface{}{
		"str": string(""),
		"num": int64(0),
		"links": []interface{}{
			&map[string]interface{}{
				"href": "/test_jobdist",
				"rel":  "index"}}}
	test_input interface{} = map[string]interface{}{
		"str": "test string in test input",
		"num": int64(118988823)}
)

type test_worker struct{}

func (tw test_worker) Work(result *map[string]interface{}) error {
	res := *result
	res["response"] = map[string]interface{}{
		"count": int64(0),
		"msg":   res["str"].(string)}

	for {
		res2 := res["response"].(map[string]interface{})
		res2["count"] = res2["count"].(int64) + 1
		if res2["count"].(int64) >= res["num"].(int64) {
			break
		}
	}
	return nil
}

func TestAll(t *testing.T) {

	var myworker test_worker

	myjob := New(test_template, test_input, myworker)

	reply := myjob.New_Form()
	fmt.Println("new_form:[", reply, "]")

	// normally would only do this if Satisfies_Template() returns true
	redir_loc := myjob.Create_Redirect()
	fmt.Println("redirect:[", redir_loc, "]")

	http.ListenAndServe("localhost:9000", nil)
}
