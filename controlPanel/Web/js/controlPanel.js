var currentHeight = 0
var leaderHeight = 0

setInterval(updateHTML,1000);

function updateHTML() {
  getHeight() // Update items related to height
}

function getHeight() {
  resp = queryState("myHeight",function(resp){
    currentHeight = parseInt(resp)
    $("#nodeHeight").val(resp)
  })

  resp = queryState("leaderHeight",function(resp){
    //$("#nodeHeight").val(resp)
    leaderHeight = parseInt(resp)
    updateProgressBar("#syncFirst > .progress-meter", currentHeight, leaderHeight)
    percent = (currentHeight/leaderHeight) * 100
    percent = Math.floor(percent)
    $('#syncFirst > .progress-meter > .progress-meter-text').text(percent + "% Synced (" + currentHeight + " of " + leaderHeight + ")")
  })

    resp = queryState("completeHeight",function(resp){
    //$("#nodeHeight").val(resp)
    completeHeight = parseInt(resp)
    updateProgressBar("#syncSecond > .progress-meter", currentHeight, completeHeight)
    percent = (completeHeight/completeHeight) * 100
    percent = Math.floor(percent)
    $('#syncSecond > .progress-meter > .progress-meter-text').text(currentHeight + " of " + leaderHeight)
  })
}

function updateProgressBar(id, current, max) {
  percent = (current/max) * 100
  $(id).width(percent+ "%")
}