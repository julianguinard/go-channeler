package channeler

import (
    "log"
)


/**
This describes the standard units that a Channeler is supposed to organize.
They are intended to be placed inside a Channeler's CallbackChain map
 */
type ChanneledCallback struct {
    //keys of Channeler.CallbackChain that need to terminate before CallbackFunction is allowed to run
    //if empty then the ChanneledCallback does not have any pre-requisite and can be ran immediately
    DependenciesNames []string
    //callbacks to execute, with various parameters number and types. Expect to return a value of various type
    //and an error. The passed parameters will be result fetched from dependenciesChannels.
    //this is this function's job to cast the interfaces mapped by variadic args appropriately
    CallbackFunction  ChanneledCallbackCallbackFunction
    /**
    => channels
        //this ChanneledCallback's dependencies channels
        => "dependencies" : type channelsMap : contains channels that may be fed by results or errors of callbacks corresponding to names defined in DependenciesNames
        //inverse of "dependencies" : channels in which this ChanneledCallback might write errors or results in order to inform other ChanneledCallback that depend on it
        => "feed" : type channelsMap : contains channels to feed with this ChanneledCallback.CallbackFunction's eventual returned result or error
     */
    channels          map[string]channelsMap
}

/**
Propagate a given message or error to channeledCallback.channels["feed"] and close each one of these channels
 */
func (channeledCallback *ChanneledCallback) propagateToFedChannels(message interface{}) {
    if (len(channeledCallback.channels["feed"]) > 0) {
        //feed result or errors to dependencies
        for fedName, fedChannel := range channeledCallback.channels["feed"] {
            log.Println("FEED CHANNEL %s WITH %s", fedName, message)
            fedChannel <- message
        }
    }
}

/**
Close each of channeler.channels and each of channeler.CallbackChain channels
 */
func (channeledCallback *ChanneledCallback) closeAllChannels() {
    for feedCallbackName, oneChannel := range channeledCallback.channels["feed"] {
        log.Println("  - Close channeledCallback's feed channel %s", feedCallbackName)
        close(oneChannel)
    }
}

/**
Initialize channeledCallback(s channels attributes to empty map
 */
func (channeledCallback *ChanneledCallback) initDependenciesFeedChannels() {
    channeledCallback.channels = map[string]channelsMap{
        "dependencies": channelsMap{},
        "feed":channelsMap{},
    }
}

/**
Initializes a new ChanneledCallback with passed public attributes and initialize the private attributes as well
 */
func NewChanneledCallback(callbackFunction ChanneledCallbackCallbackFunction, dependenciesNames []string) *ChanneledCallback {
    channeledCallback := &ChanneledCallback{
        DependenciesNames: dependenciesNames,
        CallbackFunction: callbackFunction,
    }
    return channeledCallback
}
