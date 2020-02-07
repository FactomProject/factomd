$(document).ready(function () {
    let e1 = new EventSource('/events/processlist');
    e1.onmessage = function (event) {
        console.log("processlist", event);
        $('#processlistData').text(event.data);
    };

    $(document).foundation();
});
