$(document).ready(function () {
    $(document).foundation();
    $('#nav-more').addClass('is-active');
    $('#nav-link-more').attr('aria-selected', true);

    let e1 = new EventSource('/events/processlist');
    e1.onmessage = function (event) {
        console.log("processlist", event);
        $('#processlistData').text(event.data);
    };

});
