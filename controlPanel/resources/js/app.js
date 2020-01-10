const MAX_EVENT_ITEMS = 5;

$(document).ready(function(){
    e1 = new EventSource('/events/channel-1');
    e1.onmessage = function(event) {
        console.log("channel-1", event);
        $('#channel-1').html(JSON.stringify(event.data));
    };


    var events = [];
    e2 = new EventSource('/events/channel-2');
    e2.onmessage = function(event) {
        console.log("channel-2", event);
        if (events.length >= MAX_EVENT_ITEMS) {
            events.pop();
        }
        events.unshift("item " + event.data);
        $('#channel-2').html(events.join("<br>"));
    };
});