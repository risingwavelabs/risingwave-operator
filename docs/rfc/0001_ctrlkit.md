|                    |                                                                    |
| -------            | ------------------------------------------------------------------ |
| Feature            | CtrlKit: A human-friendly framework for Kubernetes Operators       |
| Status             | Completed                                                          |
| Date               | 2022-06-14                                                         |
| Authors            | <!-- cspell:disable-line -->arkbriar                                                           |
| RFC PR #           | [#49](https://github.com/singularity-data/risingwave-operator/pull/49) |
| Implementation PR #| [#53](https://github.com/singularity-data/risingwave-operator/pull/53), [#61](https://github.com/singularity-data/risingwave-operator/pull/61), [#67](https://github.com/singularity-data/risingwave-operator/pull/67)             |
|                    |                                                                    |

# **Table of Contents**

- [**Table of Contents**](#table-of-contents)
- [**Summary**](#summary)
- [**Motivation**](#motivation)
- [**Related Resources**](#related-resources)
- [**Explanation**](#explanation)
	- [**Detailed design**](#detailed-design)
		- [**Overview**](#overview)
		- [**Concepts**](#concepts)
		- [**DSL**](#dsl)
		- [**CLI Tool**](#cli-tool)
		- [**Library**](#library)
		- [**Controller**](#controller)
		- [**Not mentioned or not related**](#not-mentioned-or-not-related)
- [**Drawbacks**](#drawbacks)
- [**Rationale and Alternatives**](#rationale-and-alternatives)
- [**Future possibilities**](#future-possibilities)

# **Summary** 

The RFC proposes a human-friendly framework `ctrlkit` for developing a robust Kubernetes operator with easy maintenance. It allows developers to break the complex reconciliation procedure into more minor and single-responsibility actions. It introduces a DSL describing the controller manager and provides a command-line tool and a library for generating polyfill codes and integrating them into projects. As a result, the developers can focus on implementing actions and organization of workflows without writing annoying polyfill codes under this framework, which leads to a more robust and maintainable system.

# **Motivation**

Operators require the developers to implement asynchronous progress to adjust the reality so that the state we observed could finally be consistent with the requirement described by custom resources. Such a way is so-called imperative. It's hard to develop and maintain an operator when its target resource is complicated. In my opinion, the problems are three-fold:
* There's no clear way to break the complex procedure into fine-grained, comprehensive units. Functions/Methods are possible but hard to reuse and organize. We need something that enforces abstraction and encapsulation.
* The control flow that PL provides isn't expressive enough. The procedure's going to be organized in procedure-style rather than graph-style.
* Too many polyfill codes to write. When we want to request subresources of another resource, we have to write a method that invokes the client of apiserver with some selectors, e.g., names and labels. Writing these polyfill codes is painful if there are many kinds of subresources.
# **Related Resources**

Some concepts and projects:

* [Kubernetes -- Custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
* [Kubernetes -- Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
* [kubernetes-sigs/controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
* [pingcap/tidb-operator](https://github.com/pingcap/tidb-operator)
* [cockroachdb/cockroach-operator](https://github.com/cockroachdb/cockroach-operator)
* [ApsaraDB/polardbx-operator](https://github.com/ApsaraDB/galaxykube)


# **Explanation**

## **Detailed design**

### **Overview**

The framework consists of 3 parts:
* A DSL to define the state and actions
* A command-line tool to generate the polyfill codes
* A library to help organize the workflow and provide utilities

### **Concepts**

**State**: Kubernetes resources obtained from apiserver represent the reality that the operator observes. Typically we query our desired resources by getting by a namespaced name or by listing with label/field selectors.

**Action**: The basic unit of workflow. The [Single Responsibility Principle](https://en.wikipedia.org/wiki/Single-responsibility_principle) should also be followed while designing the actions. Actions might take some states as input. The output of the actions affects the behavior of the event queue.

**Workflow**: A collection of actions that are organized in a graph.

### **DSL**

The DSL is for declaring the manager of some CR that contains states and actions. DSL also contains statements to help the tool generate the polyfill codes.

1. `bind GV Package`
    * `bind` a group/version (e.g., apps/v1) to a Go package (e.g., `k8s.io/api/apps/v1`), which provides the types in GV.
2. `alias Kind GVK`
    * `alias` a group/version/kind to a single word name, commonly the same name with type, to make the reference of GVK easier.
3. `decl ControllerManager for Kind` block
    * `decl` a controller manager for a specific kind, with a body as follows:
    * `state` block, with a body:
      * `name Kind` / `name []Kind` block defines a state's name and kind (where `[]Kind` means a list of resources), with a body:
        * Each line is a selector. Supported selectors and the effects are summarized in the table below.
    * `action` block, with a body:
      * `ActionName(states...)` defines an action with its name and states it requires. The `states...` part can be empty. But if there's any state required, it must be declared before in some `state` block.
4. Comments must begin with `//`.

Selector | Allow Values | Effect
:---: | :---: | :---:
name | Constant value or reference of string value of `target` | Get the resource by name (in `target`' s namespace), e.g. `name=test` or `name=${target.Name}`
labels/some-label | Constant value or reference of string value of `target` | Get the resources by listing with the labels (in `target`'s namespace), e.g. `labels/cronjob=${target.Name}`
fields/some-field | Constant value or reference of string value of `target` | Get the resources by listing with the fields (in `target`'s namespace), e.g. `fields/.metadata.name=test`. Note only some of the fields are supported on the server-side. Additional fields must be set as indices of cache in `controller-runtime` framework before it can be used.
owned | No value | The state getter will validate if the selected resources are owned by `target`. If not, the behavior differs for different states. If the state is a list, the objects not owned will be left out. If the state is a single object, an error will be returned.

An example of a document describes the states and actions needed for implementing a controller of `CronJob`:

```plain
bind v1 k8s.io/api/core/v1
bind batch/v1 k8s.io/api/batch/v1
bind controller-kit-demo/v1 controller-kit-demo/api/v1

alias CronJob controller-kit-demo/v1/CronJob
alias Job batch/v1/Job

// CronJobControllerManager declares all the actions needed by the CronJobController.
decl CronJobControllerManager for CronJob {
    state {
        jobs []Job {
            labels/cronjob=${target.Name}
            fields/.metadata.controller=${target.Name}
            owned
        }
    }

    action {
        // List all active jobs, and update the status.
        ListActiveJobsAndUpdateStatus(jobs)

        // Clean up old jobs according to the history limit.
        CleanUpOldJobsExceedsHistoryLimits(jobs)

        // Run the next job if it's on time, or otherwise we should wait 
        // until the next scheduled time.
        RunNextScheduledJob()

        // Update status of CronJob.
        UpdateCronJobStatus()
    }
}
```

### **CLI Tool**

The CLI tool takes a document file written in DSL above and generates the polyfill codes. Currently, only Go is going to be supported as the target language.

It generates the state getters and action wrappers by encapsulating them into structs and methods. It leaves the implementation of actions to the user with a stub interface. For example, it generates an output like this for the `CronJob` document above:

```go
type CronJobControllerManagerState struct {
	client.Reader
	target *apiv1.CronJob
}

func (s *CronJobControllerManagerState) GetJobs(ctx context.Context) ([]batchv1.Job, error) {
	...
}

type CronJobControllerManagerImpl interface {
	ctrlkit.ControllerManagerActionLifeCycleHook

	ListActiveJobsAndUpdateStatus(ctx context.Context, logger logr.Logger, jobs []batchv1.Job) (ctrl.Result, error)

	CleanUpOldJobsExceedsHistoryLimits(ctx context.Context, logger logr.Logger, jobs []batchv1.Job) (ctrl.Result, error)

	RunNextScheduledJob(ctx context.Context, logger logr.Logger) (ctrl.Result, error)

	UpdateCronJobStatus(ctx context.Context, logger logr.Logger) (ctrl.Result, error)
}

type CronJobControllerManager struct {
	state  CronJobControllerManagerState
	impl   CronJobControllerManagerImpl
	logger logr.Logger
}

func (m *CronJobControllerManager) ListActiveJobsAndUpdateStatus() ctrlkit.ReconcileAction {
	return ctrlkit.WrapAction("ListActiveJobsAndUpdateStatus", func(ctx context.Context) (ctrl.Result, error) {
		logger := m.logger.WithValues("action", "ListActiveJobsAndUpdateStatus")

		jobs, err := m.state.GetJobs(ctx)
		if err != nil {
			return ctrlkit.RequeueIfError(err)
		}

		defer m.impl.AfterActionRun("ListActiveJobsAndUpdateStatus", ctx, logger)
		m.impl.BeforeActionRun("ListActiveJobsAndUpdateStatus", ctx, logger)

		return m.impl.ListActiveJobsAndUpdateStatus(ctx, logger, jobs)
	})
}

...
```

The developer is responsible for writing an implementation of the `CronJobControllerManagerImpl` to implement the actions.

### **Library**

The library should contain three main parts:
* Abstraction of actions `ReconcileAction`.
* Functions for organizing the workflow `Join`, `Sequential`, `Timeout`, and `If`.
* Functions and errors for continuing or interrupting the workflow.

```go
type ReconcileAction interface {
    Description() string
	Run(ctx context.Context) (ctrl.Result, error)
}
```

**Workflow organizing functions**

Function | Example | Effect 
:---: | :---: | :---: 
`Join(actions...)`, `ParallelJoin(actions...)`, `JoinOrdered(actions...)` | `Join(Action_A, Action_B)` | Run all actions joined, and return the results joint from each action (while keep the semantic of the results). `ParallelJoin` provides parallelism besides the `Join` semantic. `JoinOrdered` guarantees the execution order.
`Sequential(actions...)` | `Sequential(Action_A, Action_B)` | Run all actions one-by-one. If any of the actions interrupts the workflow, left actions will not be executed and the result of that action will be returned.
`Timeout(timeout, action)` | `Timeout(5 * time.Second, Action_A)` | Abort the action if specified time is used.
`If(condition, action)` | `If(x == 1, Action_A)` | Action's only valid when condition is true.

`Sequential` defines a sequential flow of actions, and `Join` defines a split-join flow.

An example workflow with the actions defined above:

```go
// Run these actions and doesn't care the order, and join the results.
ctrlkit.Join(
	// Update the status of CronJob as always.
	mgr.ListActiveJobsAndUpdateStatus(),
	// Clean the old completed/failed jobs according to the limits.
	mgr.CleanUpOldJobsExceedsHistoryLimits(),
	// Try to run the next scheduled job when not suspended, otherwise do nothing.
	ctrlkit.If(cronJob.Spec.Suspend == nil || *cronJob.Spec.Suspend, mgr.RunNextScheduledJob()),
)
```

We can view it as a DAG like:

![CronJob workflow](./images/CTRLKIT%20WORKFLOW.png)


**Workflow interaction functions**

Function  | Effect 
:---: | :---: 
RequeueImmediately | Interrupts the workflow and requeue the request immediately.
RequeueAfter | Interrupts the workflow and requeue the request after given time.
RequeueIfError | Interrupts the workflow and requeue the request into queue if there's an error.
NoRequeue | No effect.
Exit | Equivalent to `RequeueIfError(ErrExit)`, where `ErrExit` is a pre-defined error to exit the workflow without reason.

**Rules of result join**

* `nil` join `err` = `err`
* `err1` join `err2` = `multierr{err1, err2}`
* `Result.Requeue`s are OR-ed.
* `Result.RequeueAfter`, we use the shortest non-zero duration if there's any, or 0 value otherwise.

These rules ensure that the interruption always happens if any action wants it and never happens later than expected. 

### **Controller**

Here's a simple example of how to implement a controller with the framework:

```go
type CronJobController struct {
	client.Client
	logr.Logger
}

func (c *CronJobController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	logger := c.Logger.WithValues("cronjob", request)

	// Get CronJob object with client.
	var cronJob apiv1.CronJob
	if err := c.Client.Get(ctx, request.NamespacedName, &cronJob); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("object not found, skip")
			return ctrlkit.NoRequeue()
		}
		logger.Error(err, "unable to get object")
		return ctrlkit.RequeueIfError(err)
	}

	// Build state and impl for controller manager.
	state := manager.NewCronJobControllerManagerState(c.Client, cronJob.DeepCopy())
	impl := manager.NewCronJobControllerManagerImpl(c.Client, cronJob.DeepCopy())
	mgr := manager.NewCronJobControllerManager(state, impl, logger)

	// Always update the status after actions have run.
	defer mgr.UpdateCronJobStatus()

	// Assemble the actions and run.
	return ctrlkit.IgnoreExit(
		// Run these actions and doesn't care the order, and join the results.
		ctrlkit.Join(
			// Update the status of CronJob as always.
			mgr.ListActiveJobsAndUpdateStatus(),
			// Clean the old completed/failed jobs according to the limits.
			mgr.CleanUpOldJobsExceedsHistoryLimits(),
			// Try to run the next scheduled job when not suspended, otherwise do nothing.
			ctrlkit.If(cronJob.Spec.Suspend == nil || *cronJob.Spec.Suspend, mgr.RunNextScheduledJob()),
		).Run(ctx),
	)
}
```

### **Not mentioned or not related**

1. How to trace the actions and workflow? We could achieve this by logging before and after each action run.
2. Problems supposed to be considered while implementing the controller:
	* When and how to observe reality? Is cache necessary? Will cache introduce additional issues?
	* When to update the target's status? How do we deal with the conflicts?

# **Drawbacks**

1. The current design is tightly tied to the `controller-runtime`.
2. You have to regenerate the stub each time the DSL doc changes.

# **Rationale and Alternatives**

Alternatives:
1. [tidb-operator](https://github.com/pingcap/tidb-operator)
   * It encapsulates the actions with functions and assembles them with control flow restricted by PL, i.e., procedure-style.
2. [cockroach-operator](https://github.com/cockroachdb/cockroach-operator)
   * It uses coarse-grained actions to perform the transitions between states. It runs only one action at the same time. IMHO, it can not respond to the new state as soon as possible. And it's hard to maintain due to the coarse-grained nature of actions.
3. [galaxykube](https://github.com/ApsaraDB/galaxykube) 
   * It is similar to the RFC proposed approach but writes pretty ugly polyfill codes. The actions are scattered among the repo and hard to use and maintain. The execution of the actions is still in the procedure-style rather than the graph-style (which is more expressive).

# **Future possibilities**

1. Enhance the DSL in such aspects:
   * Support more selectors for states.
   * Set kinds in the `core/v1`, `apps/v1`, and other common groups as built-in kinds.

2. Debug tool for visualizing the workflow of test cases.
