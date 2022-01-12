# Package goals

This Channeler package is aimed at providing an easy way to coordinate asynchronous efforts, using goroutines and channels internally.

The idea is to get the same behavior as NodeJS' async librairy auto function does, but with golang : (see https://caolan.github.io/async/v3/docs.html#auto). It allow us to parallelize execution of several functions at the same time when possible, and express blocking dependencies between these functions 

In order to do This, the package provides 2 structs:

### Channeler

This is the coordinator of the various asynchronous callbacks and has the following exported properties:
  - CallbackChain : a map of string => [ChanneledCallback](#channeledcallback) objects. This allows the Channeler to organize named ChanneledCallback objects
  - Results : map of string => interface{}, which holds the results from the callback chain described above. The map keys will match the names of the callback chain, and holds nil if an error is encountered
  - Errors : map of string => error{}, which holds the eventual errors from the callback chain described above. The map keys will match the names of the callback chain, and holds nil if no error is encountered

Each Channeler instances has a Run() method which executes the callbacks in the Channeler's CallbackChain sequentially and populate its Errors and Results properties accordingly
The module also exposes a NewChanneler() factory function which receives a CallbackChain-typed object as 1st and only argument, in order to create a Channeler instance

### ChanneledCallback

This is the unit contained by the CallbackChain map described above. It holds 2 public attributes:
  - DependenciesNames : a slice of (string) names representing keys of containing Channeler.CallbackChain that need to terminate before CallbackFunction is allowed to run 
  - CallbackFunction : a function with "func(dependencies CallbackResults) (interface{}, error)" signature. its "dependencies" parameter contains results that were needed to be fetched prior to the function's execution, as expressed by the DependenciesNames attribute. If an error is triggered during one of the dependencies' function execution, then it will be propagated and this function will not be ran as we consider the dependencies to be vital to this function's execution
  
The module exposes a NewChanneledCallback() factory method, which receives a CallbackFunction-typed object as 1st argument and DependenciesNames-typed object as 2nd 
   
