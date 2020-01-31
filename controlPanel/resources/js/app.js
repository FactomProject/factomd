const MAX_EVENT_ITEMS = 15;

$(document).ready(function () {
    let e1 = new EventSource('/events/channel-1');
    e1.onmessage = function (event) {
        console.log("channel-1", event);
        $('#channel-1').html(JSON.stringify(event.data));
    };


    let events = [];
    let generalEvents = new EventSource('/events/general-events');
    generalEvents.onmessage = function (event) {
        console.log("general events: ", event);
        if (events.length >= MAX_EVENT_ITEMS) {
            events.pop();
        }
        events.unshift(event.data);
        $('#general-events').html(events.join("<br>"));
    };

    let updateEvents = new EventSource('/events/update');
    updateEvents.onmessage = function (event) {
        data = $.parseJSON(event.data)
        console.log("update: ", data);

        $("#nodeHeight").val(data.CurrentHeight);
        $("#nodeMinute").val(data.CurrentMinute);

        $('#first-progress-meter').text(percentage(data.CurrentHeight, data.LeaderHeight) + "% Synced (" + data.CurrentHeight + " of " + data.LeaderHeight + ")");
        $('#second-progress-meter').text(percentage(data.CompleteHeight, data.LeaderHeight) + "% Synced (" + data.CompleteHeight + " of " + data.LeaderHeight + ")");

        updateProgressBar("#syncFirst > .progress-meter", data.CurrentHeight, data.LeaderHeight);
        updateProgressBar("#syncSecond > .progress-meter", data.CompleteHeight, data.LeaderHeight);
    };

    let moveToHeightEvents = [];
    let moveToHeightEventSource = new EventSource('/events/move-to-height');
    moveToHeightEventSource.onmessage = function (event) {
        // console.log("move to height even: t", data);

        if (moveToHeightEvents.length >= MAX_EVENT_ITEMS) {
            moveToHeightEvents.pop();
        }
        moveToHeightEvents.unshift(event.data);
        $('#move-to-height').html(moveToHeightEvents.join("<br>"));
    };
});

function percentage(value, ratio) {
    if (value == 0) {
        return 100;
    } else if (ratio == 0) {
        return 0;
    } else {
        return Math.floor((value / ratio) * 100);
    }
}

function updateProgressBar(id, current, max) {
    if (max == 0) {
        percent = (current / max) * 100
        $(id).width("100%")
    } else {
        percent = (current / max) * 100
        $(id).width(percent + "%")
    }
}