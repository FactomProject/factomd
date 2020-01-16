const MAX_EVENT_ITEMS = 5;

$(document).ready(function(){
    let e1 = new EventSource('/events/channel-1');
    e1.onmessage = function(event) {
        console.log("channel-1", event);
        $('#channel-1').html(JSON.stringify(event.data));
    };


    let events = [];
    let generalEvents = new EventSource('/events/general-events');
    generalEvents.onmessage = function(event) {
        console.log("general events: ", event);
        if (events.length >= MAX_EVENT_ITEMS) {
            events.pop();
        }
        events.unshift(event.data);
        $('#general-events').html(events.join("<br>"));
    };

    let moveToHeightEvents = [];
    let moveToHeightEventSource = new EventSource('/events/move-to-height');
    moveToHeightEventSource.onmessage = function(event) {
        console.log("move to height even: t", event);
        if (moveToHeightEvents.length >= MAX_EVENT_ITEMS) {
            moveToHeightEvents.pop();
        }
        moveToHeightEvents.unshift(event.data);
        $('#move-to-height').html(moveToHeightEvents.join("<br>"));
    };
});