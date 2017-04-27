package channeler

import (
    "fmt"
    //"log"
    "github.com/julianguinard/go-channeler/utils/array"
)

type variadicTypeChannel chan interface{}
//this "map type"'s key is a callback name in a Channeler's CallbackChain and the value is a channel of variable value
type channelsMap map[string]variadicTypeChannel
type CallbackResults map[string]interface{}

type CallbackChain map[string]*ChanneledCallback
type ChanneledCallbackCallbackFunction func(dependencies CallbackResults) (interface{}, error)

type DependencyError struct {
    CallbackName string
    FailedDependency string
}
func(err *DependencyError) Error() string {
    return fmt.Sprintf("%s failed dependency that must must be propagated in %s", err.CallbackName, err.FailedDependency)
}

/*
This class is intended to synchronize various ChanneledCallback objects execution by creating the
appropriate channel chain
 */
type Channeler struct {
    CallbackChain     *CallbackChain
    //========private attributes : initialized on NewChanneler() call
    //populated from CallbackChain : a channel by map entry in CallbackChain
    channels          channelsMap
    Results           CallbackResults
    Errors            map[string]error
}

/**
Factory method that initializes a Channeler and populate its attributes
 */
func NewChanneler(callbackChain *CallbackChain) *Channeler {
    channeler := &Channeler{CallbackChain:callbackChain}
    return channeler
}

func (channeler *Channeler) reset() {
    channeler.Results = CallbackResults{}
    channeler.Errors = map[string]error{}
    channeler.channels = channelsMap{}
}

func (channeler *Channeler) establishDependencyChannels() {
    channeler.reset()
    //start by resetting all dependencies channels in callback chain
    for _, channeledCallback := range *channeler.CallbackChain {
        channeledCallback.initDependenciesFeedChannels()
        //remove inexisting dependencies
    }
    for callbackName, channeledCallback := range *channeler.CallbackChain {
        channeler.channels[callbackName] = make(variadicTypeChannel, 1)
        //check if the function needs to feed channels (in case it is refered to as dependency by other ChanneledCallback
        //in the callbackChain
        for nameOfPotentiallyDependantCb, potentiallyDependantCb := range *channeler.CallbackChain {
            //exclude recursive links (self-dependencies prohibited)
            if (callbackName != nameOfPotentiallyDependantCb && (array.ArraySearchString(potentiallyDependantCb.DependenciesNames, callbackName) != -1)) {
                //log.Println(nameOfPotentiallyDependantCb + " is dependent on " + callbackName)
                //establish a dependency channel that will be fetched into current callbackFunction's feedChannels attribute,
                //and into potentiallyDependantCb's dependenciesChannels
                channeledCallback.channels["feed"][nameOfPotentiallyDependantCb] = make(variadicTypeChannel, 1)
                potentiallyDependantCb.channels["dependencies"][callbackName] = channeledCallback.channels["feed"][nameOfPotentiallyDependantCb]
                //log.Printf("EXPOSE A variadicTypeChannel DEPENDENCY BETWEEN channeledCallback.channels[\"feed\"][\"%s\"] (Adress : %s) and potentiallyDependantCb.channels[\"dependencies\"][\"%s\"] (Adress : %s)",
                    nameOfPotentiallyDependantCb,
                    channeledCallback.channels["feed"][nameOfPotentiallyDependantCb],
                    callbackName,
                    potentiallyDependantCb.channels["dependencies"][callbackName],
                )
            }
        }
    }
    //log.Println(channeler.channels)
}

/**
Close each of channeler.channels and each of channeler.CallbackChain channels
 */
func (channeler *Channeler) closeAllChannels() {
    //close the channeler's own channels
    for callbackName, oneChannel := range channeler.channels {
        //log.Printf("Close channeler's channel %s", callbackName)
        close(oneChannel)
    }
    //close the channeler's callback chain channels
    for callbackName, oneChanneledCallback := range *channeler.CallbackChain {
        if(len(oneChanneledCallback.channels["feed"]) > 0) {
            //log.Printf("- Close all feed channels for channeler's callback %s", callbackName)
            oneChanneledCallback.closeAllChannels()
        }
    }
}

/**
Launch all callbacks simultaneously (1 goroutine per callback in channeler.CallbackChain)
and block them according to their dependencies ussing their channels
 */
func (channeler *Channeler) Run() {
    channeler.establishDependencyChannels()
    for callbackName, channeledCallback := range *channeler.CallbackChain {
        go func(callbackName string, channeledCallback *ChanneledCallback, resultChannel variadicTypeChannel) {
            var err error
            var result interface{}
            dependenciesResults := CallbackResults{}
            //if there are blocking dependencies wait for them to be fetched using dependenciesChannels...
            if (len(channeledCallback.channels["dependencies"]) > 0) {
                //log.Printf("[%s] -- needs to wait for %d dependencies to be satisfied...", callbackName, len(channeledCallback.channels["dependencies"]))
                for depCbName, _ := range channeledCallback.channels["dependencies"] {
                    //log.Printf("[%s] --    - callback %s", callbackName, depCbName)
                }
                for depCbName, dependencyCbChannel := range channeledCallback.channels["dependencies"] {
                    dependenciesResults[depCbName] = <-dependencyCbChannel
                    //whenever an error is received through a dependency channel, we do not invoke the channeledCallback.CallbackFunction
                    //as the dependencies could not be fullfilled.
                    if receivedError, isOfTypeError := dependenciesResults[depCbName].(error); isOfTypeError {
                        //log.Printf("[%s] -- RECEVIED AN ERROR FROM ITS %s DEPENDENCY!! cannot call the callback function, propagate the error to feed dependencies...", callbackName, depCbName)
                        err = receivedError
                        break
                    }
                }
            }
            if(err == nil) {
                //...then call the CallbackFunction along with the args from dependencies if any...
                result, err = channeledCallback.CallbackFunction(dependenciesResults)
                //log.Printf("[%s] -- HAS RETURNED result %s and error %s", callbackName, result, err)
            }

            if (err == nil) {
                channeledCallback.propagateToFedChannels(result)
                resultChannel <- result
            } else {
                channeledCallback.propagateToFedChannels(err)
                resultChannel <- err
            }
            //log.Printf("============= END OF GOROUTINE %s==========================", callbackName)
        }(callbackName, channeledCallback, channeler.channels[callbackName])
    }

    //wait for all results or errors to be gathered...
    for callbackName, callbackChannel := range channeler.channels {
        finalReceived := <-callbackChannel
        //cast errors if received as we go
        if receivedError, isOfTypeError := finalReceived.(error); isOfTypeError {
            channeler.Errors[callbackName] = receivedError
            channeler.Results[callbackName] = nil
            //log.Printf("FINAL %s===> RECEIVED ERROR %s", callbackName, channeler.Errors[callbackName])
        } else {
            channeler.Results[callbackName] = finalReceived
            channeler.Errors[callbackName] = nil
            //log.Printf("FINAL %s===> RECEIVED RESULT %s", callbackName, finalReceived)
        }
    }
    channeler.closeAllChannels()
    //log.Printf("===================== ALL DONE, results %s =======================", channeler.Results)
    //...from now on then all results must be accessible from channeler.Results
}
