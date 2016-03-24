package jobdist

import (
	"encoding/json"
	"fmt"
	"github.com/dankozitza/dkutils"
	"github.com/dankozitza/sconf"
	"github.com/dankozitza/stattrack"
	"net/http"
)

var (
	Job_File = "jobs.json"
	stat     = stattrack.New("package initialized")
	job_cnt  = int64(0)
	jobdb    = sconf.New( // not in use
		Job_File,
		sconf.Sconf{
			"max_jobs": int64(100),
			"jobs":     map[string]interface{}{}})
)

// Worker
//
// An interface that must be provided to Job.New().
//
type Worker interface {

	// Work
	//
	// The Work function is called from Create_Redirect(). When a Job's template
	// is satisfied by the input parameter passed into Job.New() a Job can call
	// Create_Redirect() which will populate the Job.Response map with the
	// contents of Job.Input, the status, and a links array. It then calls
	// Work(Job.Response). Work is expected to use the input parameters specified
	// in Job.Template to populate (*result)["response"] with whatever resource
	// the job provides.
	//
	Work(result *map[string]interface{}) error
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

// Satisfies_Template
//
// Attempts to convert input parameter type to template type recursively. If
// this returns false then Jobs.Input may not be the type expected by the
// worker and could cause a panic in the Work() function.
//
func (j *Job) Satisfies_Template() bool {

	// here i could copy the links array from the template to the input
	// so that dkutils.DeepTypeCheck would not need to

	result, err := dkutils.DeepTypePersuade(j.Template, j.Input)
	if err == nil {
		j.Input = result
		return true
	}
	stat.Warn("Satisfies_Template: got error: " + err.Error())
	return false
}

// New_Form()
//
// When the template is not satisfied by the input parameter reply with this
// form generated from the template.
//
func (j *Job) New_Form() interface{} {
	var ret interface{} = map[string]interface{}{"job": j.Template}
	r_map := ret.(map[string]interface{})
	j_map := r_map["job"].(map[string]interface{})
	j_map["status"] = "awaiting_input"
	return ret
}

// Create_Redirect
//
// Prepares Job.Response to be passed into (*Job.worker).Work(), creates HTTP
// handler for the redirect, registers the handler with http using the generated
// href, executes (*Job.worker).Work() in a seperate go routine, and returns
// the generated href.
//
func (j *Job) Create_Redirect() string {

	*j.Status = "in_progress"

	redir_href := "/jobs/" + fmt.Sprint(j.Id)

	jhh := JobHTTPHandler(*j)
	http.Handle(redir_href, jhh)

	// add the input to the response
	*j.Response = j.Input.(map[string]interface{})
	(*j.Response)["status"] = &j.Status
	// add the self rel to the links slice
	(*j.Response)["links"] = []interface{}{
		&map[string]interface{}{
			"href": redir_href,
			"rel":  "self"}}
	// add the index rel from the template to the links slice
	tlinks := (j.Template.(map[string]interface{}))["links"].([]interface{})
	for _, v := range tlinks {
		if (*v.(*map[string]interface{}))["rel"] == "index" {
			(*j.Response)["links"] = append(
				(*j.Response)["links"].([]interface{}), v)
		}
	}
	// create the response map
	(*j.Response)["response"] = new(map[string]interface{})

	// TODO: handle error returned by do_work
	go j.do_work()

	return redir_href
}

// do_work
//
// Called by Create_Redirect() to set status after Work() is done.
//
func (j *Job) do_work() error {

	err := (*j.worker).Work(j.Response)
	*j.Status = "finished"

	return err
}

// JobHTTPHandler
//
// Handler used to generate a resource for a Job.
//
type JobHTTPHandler Job

func (jhh JobHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	m_map, err := json.MarshalIndent(
		map[string]interface{}{"job": jhh.Response},
		"",
		"   ")
	if err != nil {
		stat.PanicErr("could not marshal Job.Response", err)
	}
	fmt.Fprint(w, string(m_map))
}
