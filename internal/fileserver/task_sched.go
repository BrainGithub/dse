package fileserver

import (
    "dse/internal/fileserver/service/fsvrmgr"
    "log"
    "sync"
    "time"
)

const (
    JobSchedNum      = 2
    JobQueueCap      = 100
    JobRetryNum      = 6
    JobRetryInterval = 10 * time.Second
    InQueueTimeout   = 30 * time.Second
)

// Job is the job object of workpool.
type Job struct {
    callback jobCB
    fsvr     fsvrmgr.FileSvr
    retryCnt int
}

type jobCB func(fsvr fsvrmgr.FileSvr) error

var jobs = make(chan Job, JobQueueCap)
var jobsRetry = make(chan Job, JobQueueCap)
var jobsMap = map[int]int{}
var retryJobsMap = map[int]int{}
var jobMutex = &sync.Mutex{}
var retryMutex = &sync.Mutex{}

func sizeSchedulerInit() {
    go createWorkerPool(JobSchedNum)
    go createRetryWorkerPool(1)
}

// need to start a go routine
func createWorkerPool(n int) {
    for i := 0; i < n; i++ {
        go worker()
    }
}

func createRetryWorkerPool(n int) {
    for i := 0; i < n; i++ {
        go retryWorker()
    }
}

func worker() {
    for {
        select {
        case job := <-jobs:
            if job.Do() != nil {
                log.Printf("failed to call jobs: %v.\n", err)
            }

            jobMutex.Lock()
            delete(jobsMap, job.fsvr.Tid)
            jobMutex.Unlock()
        }
    }
}

func retryWorker() {
    for {
        select {
        case job := <-jobsRetry:
            if job.Do() != nil {
                log.Printf("failed to call jobs: %v.\n", err)
            }

            retryMutex.Lock()
            delete(retryJobsMap, job.fsvr.Tid)
            retryMutex.Unlock()
        }
    }
}

func (j Job) Do() error {
    ok, err := j.fsvr.FileFerry()
    log.Printf("retry:%d, do job:%+v, ok:%v, err:%+v", j.retryCnt, j, ok, err)

    if err != nil {
        if j.retryCnt >=6 {
            log.Printf("retry to max time:%d", j.retryCnt)
            return err
        }
        select {
        case <- time.After(JobRetryInterval):
            addShedRetryJob(j)
        }
    }
    return nil
}

func addShedJob(fsvr fsvrmgr.FileSvr) (ok bool) {
    job := Job{fsvr:fsvr}

    jobMutex.Lock()
    if jobsMap[job.fsvr.Tid] > 0 {
        jobsMap[job.fsvr.Tid]++
        jobMutex.Unlock()
        return
    }

    select {
    case jobs <- job:
        jobMutex.Lock()
        jobsMap[job.fsvr.Tid]++
        jobMutex.Unlock()
        log.Printf("job:%+v add to work queue success", job)
        return true
    case <-time.After(InQueueTimeout):
        log.Printf("job:%+v add to work queue failed", job)
        return false
    }
}

func addShedRetryJob(job Job) (ok bool) {
    retryMutex.Lock()
    if retryJobsMap[job.fsvr.Tid] > 0 {
        retryJobsMap[job.fsvr.Tid]++
        retryMutex.Unlock()
        log.Printf("ignore ticket id:%d, times:%d", job.fsvr.Tid, retryJobsMap[job.fsvr.Tid])
        return
    }
    retryMutex.Unlock()

    job.retryCnt++

    select {
    case jobsRetry <- job:
        retryMutex.Lock()
        retryJobsMap[job.fsvr.Tid]++
        retryMutex.Unlock()
        log.Printf("job:%+v add to work queue success", job)
        return true
    case <-time.After(InQueueTimeout):
        job.retryCnt--
        log.Printf("job:%+v add to work queue failed", job)
        return false
    }
}
