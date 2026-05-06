# Project \#3: Your Choice!

**See gradescope for due date**

## Assignment

The final project gives you the opportunity to show me what you learned
in this course and to build your own parallel system. In particular, you
should think about implementing a parallel system in the domain you are
most comfortable in (data science, machine learning, computer graphics,
etc.). The system should solve a problem that can benefit from some form
of parallelization and can be implemented in the way specified below.
I recommend reading the entire description before deciding what to implement.
If you are having trouble coming up with a problem for your system to
solve then consider the following:

-   [Embarrassingly Parallel
    Topics](https://en.wikipedia.org/wiki/Embarrassingly_parallel)
-   [Parallel
    Algorithms](https://en.wikipedia.org/wiki/Parallel_computing#Algorithmic_methods)

You are free to implement any parallel algorithm you like. However, you
are required to at least have the following features in your parallel
system:

-   An input/output component that allows the program to read in data or
    receive data in some way. The system will perform some
    computation(s) on this input and produce an output result.

-   A sequential implementation of the system. Make sure to provide a
    usage statement.

-   Basic Parallel Implementation: An implementation that uses the BSP
    pattern (using a condition variable to implement the barrier between
    supersteps), **or** a pipelining pattern (using channels) **or** a
    map-reduce implementation (again using a condition variable as barrier
    between the map and the reduce stage). Choose whichever is most suitable
    for solving the problem you have decided to tackle. The work in each
    stage or superstep should be divided among threads in a simple fashion.
    For example, if you choose an image processing problem with N images,
    then each of your T threads might be assigned to work on approximately
    N/T images. The easiest and most reasonable way to divide the work will
    depend on your problem and your chosen parallelization approach.

-   Work-stealing refinement: A work-stealing algorithm using a **dequeue**
    should be used such that the work can be split into smaller tasks, which
    are placed in a work queue such that threads will steal work from other threads
    when idle. You have some freedom in how exactly you want to implement the dequeue;
    it could either be based on an array as shown in class, where the owner pops items
    from one end whereas thieves pop items from the other end, or it could be
    implemented as a conventional dequeue where one end is used for enqueing, and the
    other end is only used for dequeueing, similar to the CIVL implementation. You can
    also choose other implementations as long as you justify why they make sense in your
    report. The one major requirement is that **your data structure is lock-free**.

-   Provide a detailed write-up and analysis of your system. For this
    assignment, this write-up is required to have more detail to explain
    your parallel implementations since we are not giving you a problem
    to solve. See the **System Write-up** section for more details.

-   Provide all the dataset files you used in your analysis portion of
    your write up. If these files are too big then you need to provide us
    a link so we can easily download them from an external source.
    It is likely that the work-stealing refinement is only beneficial if your
    input data is structured in a certain way, e.g. if items in the input are of vastly
    different sizes, or if subtasks in your algorithm have varying or unpredictable costs.
    Make sure that this is the case for your project, so that you can showcase the pros/cons of all implementations.

-   The grade also include design points. You should think about the
    modularity of the system you are creating. Think about splitting
    your code into appropriate packages, when necessary.

-   **You must provide a script or specific commands that shows/produces
    the results of your system**. We need to be able to enter in a
    single command in the terminal window and it will run and produce
    the results of your system. Failing to provide a straight-forward
    way of executing your system that produces its result will result in
    **significant deductions** to your score. We prefer running a simple
    command line script (e.g., shell-script or python3 script). However,
    providing a few example cases of possible execution runs will be
    acceptable.

-   We should also be able to run specific versions of the system. There
    should be an option (e.g. via command line argument) to run the
    sequential version, or the various parallel versions. Please make
    sure to document this in your report or via the printing of a usage
    statement.

-   You are free to use any additional standard/third-party libraries as
    you wish. However, all the parallel work is **required** to be
    implemented by you.

-   There is a directory called `proj3` with a single `go.mod` file
    inside your repositories. Place all your work for project 3 inside
    this directory.

### System Write-up

In prior assignments, we provided you with the input files or data to
run experiments against a your system and provide an analysis of those
experiments. For this project, you will do the same with the exception
that you will produce the data needed for your experiments. In all, you
should do the following for the writeup:

-   Run experiments with data you generate for both the sequential and
    parallel versions. For
    the parallel version, make sure you are running your experiments
    with at least producing work for `N` threads, where
    `N = {2,4,6,8,12}`. 
    
    **Cluster Requirement**: Please run your final experiments on the cluster node (`fe.ai.cs.uchicago.edu` `peanut-cpu` partition). You can find detail instruction on Canvas.
-   Produce speedup graph(s) for those data sets. You should have one
    speedup graph per parallel implementation you define in your system.

Please submit a report (pdf document, text file, etc.) summarizing your
results from the experiments and the conclusions you draw from them.
Your report should include your plot(s) as specified above and a
self-contained report. That is, somebody should be able to read the
report alone and understand what code you developed, what experiments
you ran and how the data supports the conclusions you draw. The report
**must** also include the following:

-   Describe your program and the problem it is trying to solve in detail.
-   A description of how you implemented your parallel solutions, and why
    the approach you picked (BSP, map-reduce, pipelining) is the most appropriate. You probably
    want to discuss things like load balancing, latency/throughput, etc.
-   Describe the challenges you faced while implementing the system.
    What aspects of the system might make it difficult to parallelize?
    In other words, what did you hope to learn by doing this assignment?
-   Did the usage of a task queue with work stealing improve performance?
    Why or why not?
-   What are the **hotspots** (i.e., places where you can parallelize
    the algorithm) and **bottlenecks** (i.e., places where there is
    sequential code that cannot be parallelized) in your sequential
    program? Were you able to parallelize the hotspots and/or remove the
    bottlenecks in the parallel version?
-   What limited your speedup? Is it a lack of parallelism?
    (dependencies) Communication or synchronization overhead? As you try
    and answer these questions, we strongly prefer that you provide data
    and measurements to support your conclusions.
-   Compare and contrast the two parallel implementations. Are there
    differences in their speedups?

## Don't know What to Implement?

If you are unsure what to implement then by default you can reimplement
the image processing assignment using the required new features.

**You cannot reimplement project 2 or other assignments**.

## Design, Style and Cleaning up

Before you submit your final solution, you should, remove

-   any `Printf` statements that you added for debugging purposes and
-   all in-line comments of the form: "YOUR CODE HERE" and "TODO ..."
-   Think about your function decomposition. No code duplication. This
    homework assignment is relatively small so this shouldn't be a major
    problem but could be in certain problems.

Go does not have a strict style guide. However, use your best judgment
from prior programming experience about style. Did you use good variable
names? Do you have any lines that are too long, etc.

As you clean up, you should periodically save your file and run your
code through the tests to make sure that you have not broken it in the
process.

## Grading

For this project, we grade as follows:
 - 50% Completeness. Your code should implement the required features without deadlocks or race conditions.
 - 20% Performance. Does your code scale, did you avoid unnecessary data copies, did you make an effort to remove obvious performance bottlenecks.
 - 20% Writeup. Is the report detailed, reasonably well written, and contains all the parts we asked for.
 - 10% Design and Style.

## Submission

Before submitting, make sure you've added, committed, and pushed all
your code to GitHub. You must submit your final work through Gradescope
(linked from our Canvas site) in the "Project \#3" assignment page via
two ways,

1.  **Uploading from Github directly (recommended way)**: You can link
    your Github account to your Gradescope account and upload the
    correct repository based on the homework assignment. When you submit
    your homework, a pop window will appear. Click on "Github" and then
    "Connect to Github" to connect your Github account to Gradescope.
    Once you connect (you will only need to do this once), then you can
    select the repsotiory you wish to upload and the branch (which
    should always be "main" or "master") for this course.
2.  **Uploading via a Zip file**: You can also upload a zip file of the
    homework directory. Please make sure you upload the entire directory
    and keep the initial structure the **same** as the starter code;
    otherwise, you run the risk of not passing the automated tests.

As a reminder, for this assignment, there will be **no autograder** on
Gradescope. We will run the program the CS Peanut cluster and manually
enter in the grading into Gradescope. However, you **must still submit
your final commit to Gradescope**.

