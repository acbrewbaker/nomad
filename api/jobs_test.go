package api

import (
	"reflect"
	"strings"
	"testing"
)

func TestJobs_Register(t *testing.T) {
	c, s := makeClient(t, nil, nil)
	defer s.Stop()
	jobs := c.Jobs()

	// Listing jobs before registering returns nothing
	resp, qm, err := jobs.List(nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if qm.LastIndex != 0 {
		t.Fatalf("bad index: %d", qm.LastIndex)
	}
	if n := len(resp); n != 0 {
		t.Fatalf("expected 0 jobs, got: %d", n)
	}

	// Create a job and attempt to register it
	job := testJob()
	eval, wm, err := jobs.Register(job, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if eval == "" {
		t.Fatalf("missing eval id")
	}
	assertWriteMeta(t, wm)

	// Query the jobs back out again
	resp, qm, err = jobs.List(nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	assertQueryMeta(t, qm)

	// Check that we got the expected response
	if len(resp) != 1 || resp[0].ID != job.ID {
		t.Fatalf("bad: %#v", resp[0])
	}
}

func TestJobs_Info(t *testing.T) {
	c, s := makeClient(t, nil, nil)
	defer s.Stop()
	jobs := c.Jobs()

	// Trying to retrieve a job by ID before it exists
	// returns an error
	_, _, err := jobs.Info("job1", nil)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got: %#v", err)
	}

	// Register the job
	job := testJob()
	_, wm, err := jobs.Register(job, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	assertWriteMeta(t, wm)

	// Query the job again and ensure it exists
	result, qm, err := jobs.Info("job1", nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	assertQueryMeta(t, qm)

	// Check that the result is what we expect
	if result == nil || result.ID != job.ID {
		t.Fatalf("expect: %#v, got: %#v", job, result)
	}
}

func TestJobs_Allocations(t *testing.T) {
	c, s := makeClient(t, nil, nil)
	defer s.Stop()
	jobs := c.Jobs()

	// Looking up by a non-existent job returns nothing
	allocs, qm, err := jobs.Allocations("job1", nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if qm.LastIndex != 0 {
		t.Fatalf("bad index: %d", qm.LastIndex)
	}
	if n := len(allocs); n != 0 {
		t.Fatalf("expected 0 allocs, got: %d", n)
	}

	// TODO: do something here to create some allocations for
	// an existing job, lookup again.
}

func TestJobs_Evaluations(t *testing.T) {
	c, s := makeClient(t, nil, nil)
	defer s.Stop()
	jobs := c.Jobs()

	// Looking up by a non-existent job ID returns nothing
	evals, qm, err := jobs.Evaluations("job1", nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if qm.LastIndex != 0 {
		t.Fatalf("bad index: %d", qm.LastIndex)
	}
	if n := len(evals); n != 0 {
		t.Fatalf("expected 0 evals, got: %d", n)
	}

	// Insert a job. This also creates an evaluation so we should
	// be able to query that out after.
	job := testJob()
	evalID, wm, err := jobs.Register(job, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	assertWriteMeta(t, wm)

	// Look up the evaluations again.
	evals, qm, err = jobs.Evaluations("job1", nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	assertQueryMeta(t, qm)

	// Check that we got the evals back
	if n := len(evals); n == 0 || evals[0].ID != evalID {
		t.Fatalf("expected 1 eval (%s), got: %#v", evalID, evals)
	}
}

func TestJobs_Delete(t *testing.T) {
	c, s := makeClient(t, nil, nil)
	defer s.Stop()
	jobs := c.Jobs()

	// Register a new job
	job := testJob()
	_, wm, err := jobs.Register(job, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	assertWriteMeta(t, wm)

	// Attempting delete on non-existing job does not error
	wm2, err := jobs.Delete("nope", nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	assertWriteMeta(t, wm2)

	// Deleting an existing job works
	wm3, err := jobs.Delete("job1", nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	assertWriteMeta(t, wm3)

	// Check that the job is really gone
	result, qm, err := jobs.List(nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	assertQueryMeta(t, qm)
	if n := len(result); n != 0 {
		t.Fatalf("expected 0 jobs, got: %d", n)
	}
}

func TestJobs_ForceEvaluate(t *testing.T) {
	c, s := makeClient(t, nil, nil)
	defer s.Stop()
	jobs := c.Jobs()

	// Force-eval on a non-existent job fails
	_, _, err := jobs.ForceEvaluate("job1", nil)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got: %#v", err)
	}

	// Create a new job
	_, wm, err := jobs.Register(testJob(), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	assertWriteMeta(t, wm)

	// Try force-eval again
	evalID, wm, err := jobs.ForceEvaluate("job1", nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	assertWriteMeta(t, wm)

	// Retrieve the evals and see if we get a matching one
	evals, qm, err := jobs.Evaluations("job1", nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	assertQueryMeta(t, qm)
	for _, eval := range evals {
		if eval.ID == evalID {
			return
		}
	}
	t.Fatalf("evaluation %q missing", evalID)
}

func TestJobs_NewBatchJob(t *testing.T) {
	job := NewBatchJob("job1", "myjob", 5)
	expect := &Job{
		ID:       "job1",
		Name:     "myjob",
		Type:     JobTypeBatch,
		Priority: 5,
	}
	if !reflect.DeepEqual(job, expect) {
		t.Fatalf("expect: %#v, got: %#v", expect, job)
	}
}

func TestJobs_NewServiceJob(t *testing.T) {
	job := NewServiceJob("job1", "myjob", 5)
	expect := &Job{
		ID:       "job1",
		Name:     "myjob",
		Type:     JobTypeService,
		Priority: 5,
	}
	if !reflect.DeepEqual(job, expect) {
		t.Fatalf("expect: %#v, got: %#v", expect, job)
	}
}

func TestJobs_SetMeta(t *testing.T) {
	job := &Job{Meta: nil}

	// Initializes a nil map
	out := job.SetMeta("foo", "bar")
	if job.Meta == nil {
		t.Fatalf("should initialize metadata")
	}

	// Check that the job was returned
	if job != out {
		t.Fatalf("expect: %#v, got: %#v", job, out)
	}

	// Setting another pair is additive
	job.SetMeta("baz", "zip")
	expect := map[string]string{"foo": "bar", "baz": "zip"}
	if !reflect.DeepEqual(job.Meta, expect) {
		t.Fatalf("expect: %#v, got: %#v", expect, job.Meta)
	}
}

func TestJobs_Constrain(t *testing.T) {
	job := &Job{Constraints: nil}

	// Create and add a constraint
	out := job.Constrain(HardConstraint("kernel.name", "=", "darwin"))
	if n := len(job.Constraints); n != 1 {
		t.Fatalf("expected 1 constraint, got: %d", n)
	}

	// Check that the job was returned
	if job != out {
		t.Fatalf("expect: %#v, got: %#v", job, out)
	}

	// Adding another constraint preserves the original
	job.Constrain(SoftConstraint("memory.totalbytes", ">=", "128000000", 2))
	expect := []*Constraint{
		&Constraint{
			Hard:    true,
			LTarget: "kernel.name",
			RTarget: "darwin",
			Operand: "=",
			Weight:  0,
		},
		&Constraint{
			Hard:    false,
			LTarget: "memory.totalbytes",
			RTarget: "128000000",
			Operand: ">=",
			Weight:  2,
		},
	}
	if !reflect.DeepEqual(job.Constraints, expect) {
		t.Fatalf("expect: %#v, got: %#v", expect, job.Constraints)
	}
}