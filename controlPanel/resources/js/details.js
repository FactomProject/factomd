$(document).ready(function () {
    $(document).foundation();
    $('#nav-more').addClass('is-active');
    $('#nav-link-more').attr('aria-selected', true);


    let eventSource = new EventSource('/events/summary');
    eventSource.onmessage = function (event) {
        $('#summaryData').text(event.data);
    };

    $('#tabs_details').on('change.zf.tabs', function() {
        if($('#summary:visible').length){
            eventSource.close();
            eventSource = new EventSource('/events/summary');
            eventSource.onmessage = function (event) {
                $('#summaryData').text(event.data);
            };
        }
        if($('#processlist:visible').length){
            eventSource.close();
            eventSource = new EventSource('/events/processlist');
            eventSource.onmessage = function (event) {
                $('#processlistData').text(event.data);
            };
        }
        if($('#printmap:visible').length){
            eventSource.close();
            eventSource = new EventSource('/events/printmap');
            eventSource.onmessage = function (event) {
                $('#printmapData').text(event.data);
            };
        }
    });

    $(".fullscreen-option").click( function(){
        console.log("goto fullscreen");
        txtArea = $(this).siblings(".is-active");
        txtArea.toggleClass("fullscreen");
        $(this).toggleClass("absolute-fullscreen-option");
        $(this).toggleClass("fixed-fullscreen-option");
    })

});
