var currentHeight = 0
var leaderHeight = 0

setInterval(updateHTML,1000);

function updateHTML() {
  getHeight() // Update items related to height
  updateTransactions()
}

function updateTransactions() {
  resp = queryState("recentTransactions",function(resp){
    obj = JSON.parse(resp)
    console.log(obj)
    if($("#DBKeyMR > a").text() != obj.DirectoryBlock.KeyMR) {
      $("#DBKeyMR > a").text(obj.DirectoryBlock.KeyMR)
      $("#DBBodyKeyMR").text(obj.DirectoryBlock.BodyKeyMR)
      $("#DBFullHash").text(obj.DirectoryBlock.FullHash)
      $("#DBBlockHeight").text(obj.DirectoryBlock.DBHeight)

      console.log(obj)

      $("#panFactoids > #traxList > tbody").html("")
      obj.FactoidTransactions.forEach(function(trans) {
        if(trans.TotalInput > 0.0001) {
          $("#panFactoids > #traxList > tbody").append("\
          <tr>\
              <td><a id='factom-search-link' type='facttransaction'>" + trans.TxID + "</a></td>\
              <td>" + trans.TotalInput + "</td>\
              <td>" + trans.TotalInputs + "</td>\
              <td>" + trans.TotalOutputs + "</td>\
          </tr>")
        }
      })

      $("#panEntries > #traxList > tbody").html("")
      if(obj.Entries != null){
        obj.Entries.forEach(function(entry) {
          $("#panEntries > #traxList > tbody").append("\
          <tr>\
              <td><a id='factom-search-link' type='entry'>" + entry.Hash + "</a></td>\
              <td><a id='factom-search-link' type='chainhead'>" + entry.ChainID  + "</a></td>\
              <td>" + entry.ContentLength + "</td>\
          </tr>")
        })
      }

      $("section #factom-search-link").click(function() {
        type = jQuery(this).attr("type")
        hash = jQuery(this).text()
        var x = new XMLHttpRequest()
        x.onreadystatechange = function() {
          if(x.readyState == 4) {
            obj = JSON.parse(x.response)
            if (obj.Type != "None") {
              window.location = "search?input=" + hash + "&type=" + type
              //redirect("search?input=" + hash + "&type=" + type, "post", x.response) // Something found
            } else {
              $(".factom-search-error").slideDown(300)
              console.log(x.response)
            }
          }
        }
        var formDataLink = new FormData();
        formDataLink.append("method", "search")
        formDataLink.append("search", hash)

        x.open("POST", "/post")
        x.send(formDataLink)
      })
    }
  })
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