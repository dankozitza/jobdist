package jobdist

import (
	"fmt"
	"net/http"
	"testing"
)

var (
	test_template interface{} = map[string]interface{}{
		"str": "",
		"num": 0}
	test_input interface{} = map[string]interface{}{
		"str": "test string in test input",
		"num": 7}

	//testJob Job = New(test_template, test_input, test_worker)
)

type test_worker struct{}

func (tw test_worker) Work(result *map[string]interface{}) error {
	res := *result
	//res["response"] = res["str"].(string) + ":" + fmt.Sprint(res["num"])
	res["response"] = 1
	for {
		res["response"] = res["response"].(int) + 1
		if res["response"].(int) >= 41898882 {
			break
		}
	}
	//tw.Status = "finished"
	return nil
}

func TestAll(t *testing.T) {

	for i := 0; i < 5; i++ {
		var myworker test_worker

		myjob := New(test_template, test_input, myworker)

		if !myjob.Satisfies_Template() {
			reply := myjob.New_Form()
			fmt.Println("input does not satisfy template")
			fmt.Println("new_form: ", reply)
		} else {
			fmt.Println("input satisfies template")
			//myjob.worker = myjob.input.(map[string]interface{})
			redir_loc := myjob.Create_Redirect()
			fmt.Println("you are being redirected to:", redir_loc)
		}
	}

	http.ListenAndServe("localhost:9000", nil)
}
