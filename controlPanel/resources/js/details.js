$(document).ready(function () {
    $(document).foundation();
    $('#nav-more').addClass('is-active');
    $('#nav-link-more').attr('aria-selected', true);

    let eventSource = new EventSource('/events/summary');
    eventSource.onmessage = function (event) {
        $('#summaryData').html(event.data);
    };

    $('#tabs_details').on('change.zf.tabs', function() {
        if($('#summary:visible').length){
            eventSource.close();
            eventSource = new EventSource('/events/summary');
            eventSource.onmessage = function (event) {
                $('#summaryData').html(event.data);
            };
        }
        if($('#processlist:visible').length){
            eventSource.close();
            eventSource = new EventSource('/events/processlist');
            eventSource.onmessage = function (event) {
                $('#processlistData').html(event.data);
            };
        }
        if($('#printmap:visible').length){
            eventSource.close();
            eventSource = new EventSource('/events/printmap');
            eventSource.onmessage = function (event) {
                $('#printmapData').html(event.data);
            };
        }
        if($('#servers:visible').length){
            eventSource.close();
            eventSource = new EventSource('/events/servers');
            eventSource.onmessage = function (event) {
                let data = $.parseJSON(event.data);
                console.log("servers event: ", data);
                $('#authorities').html(data.Authorities);
                $('#identities').html(data.Identities);
                $('#node').html(data.Node);
            };
        }
    });

    $(".fullscreen-option").click( function(){
        txtArea = $(this).siblings(".is-active");
        txtArea.toggleClass("fullscreen");
        $(this).toggleClass("absolute-fullscreen-option");
        $(this).toggleClass("fixed-fullscreen-option");
    });
});
