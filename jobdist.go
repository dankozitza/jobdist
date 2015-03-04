package jobdist

import (
	"encoding/json"
	"fmt"
	"github.com/dankozitza/sconf"
	"github.com/dankozitza/stattrack"
	//"github.com/dankozitza/dkutils"
	"net/http"
)

var (
	Job_File = "jobs.json"
	stat     = stattrack.New("package initialized")
	job_cnt  = int64(0)
	jobdb    = sconf.New(
		Job_File,
		sconf.Sconf{
			"max_jobs": int64(100),
			"jobs":     map[string]*Job{}})
)

type Worker interface {
	Work(result *map[string]interface{}) error // may need an input parameter as well
}

type Job struct {
	Id       int64
	Status   *string
	Template interface{}
	Input    interface{}
	Response *map[string]interface{}
	worker   *Worker
}

func New(template interface{}, input interface{}, worker Worker) *Job {

	newjob := &Job{
		job_cnt,
		new(string),
		template,
		input,
		new(map[string]interface{}),
		&worker}

	job_cnt++

	return newjob
	//jobdb["jobs"][string(jobdb["job_cnt"])] = newjob
}

func (j *Job) Satisfies_Template() bool {
	//result, err := dkutils.DeepPersuadeType(j.Template, j.Input)
	//if err == nil {
	//	//j.Input = result
	return true
	//}
	//return false
}

func (j *Job) New_Form() interface{} {
	var ret interface{} = map[string]interface{}{"job": j.Template}
	r_map := ret.(map[string]interface{})
	j_map := r_map["job"].(map[string]interface{})
	j_map["status"] = "awaiting_input"
	return ret
}

func (j *Job) Create_Redirect() string {

	*j.Status = "in_progress"

	jhh := JobHTTPHandler(*j)
	http.Handle("/jobs/"+fmt.Sprint(j.Id), jhh)

	go j.do_work()

	//(*j.worker).Work(j.Response)
	//fmt.Println("finished work")

	//*j.Status = "finished"

	return "/jobs/" + fmt.Sprint(j.Id)
}

func (j *Job) do_work() error {

	//for k, v := range j.Input.(map[string]interface{}) {
	//	(*j.Response)[k] = v
	//}

	*j.Response = j.Input.(map[string]interface{})
	(*j.Response)["status"] = &j.Status
	(*j.Response)["response"] = new(map[string]interface{})
	(*j.worker).Work(j.Response)
	*j.Status = "finished"
	fmt.Println("finished work")

	return nil
}

type JobHTTPHandler Job

func (jhh JobHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//var ret interface{} = map[string]interface{}{"job": jhh.Input}
	//r_map := ret.(map[string]interface{})
	//j_map := r_map["job"].(map[string]interface{})
	//j_map["status"] = jhh.Status

	//for k, v := range jhh.Input.(map[string]interface{}) {
	//	j_map[k] = v
	//}

	//j_map["response"] = jhh.Response

	//(*jhh.Response)["status"] = jhh.Status

	m_map, err := json.MarshalIndent(jhh.Response, "", "   ")
	if err != nil {
		stat.PanicErr("could not marshal Job.Response", err)
	}
	fmt.Fprint(w, string(m_map))
}
