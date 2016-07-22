var currentHeight = 0
var leaderHeight = 0

setInterval(updateHTML,1000);
var serverOnline = false


function updateHTML() {
  $.ajax('/', {
    success: function(){
      serverOnline = true
    },
    error: function(){
      serverOnline = false
    }
  });

  if (!serverOnline) {
    $("#server-status").text("Factomd Not Running")
    return
  } else {
    $("#server-status").text("Factomd Running")
  }
  getHeight() // Update items related to height
  updateTransactions()
  updataDataDumps()
  updatePeers()
}

$("#dump-container #fullscreen-option").click( function(){
  txtArea = jQuery(this).siblings(".is-active")
  txtArea.toggleClass("fullscreen")
  jQuery(this).toggleClass("absolute-fullscreen-option")
  jQuery(this).toggleClass("fixed-fullscreen-option")
})


// Top tabs on index
$("#indexnav-main > a").click(function() {
  if (jQuery(this).hasClass("is-active")) {

  } else {
    $("#transactions").removeClass("hide")
    $("#local").removeClass("hide")
    $("#dataDump").addClass("hide")
  }
})

$("#indexnav-more > a").click(function() {
  if (jQuery(this).hasClass("is-active")) {

  } else {
    $("#transactions").addClass("hide")
    $("#local").addClass("hide")
    $("#dataDump").removeClass("hide")
  }
})

function updataDataDumps() {
  resp = queryState("dataDump",function(resp){
    obj = JSON.parse(resp)
    $("#dump1 #dumpShort").text(obj.DataDump1.ShortDump)
    $("#dump1 #dumpRaw").text(obj.DataDump1.RawDump)

    $("#dump2 #dumpRaw").text(obj.DataDump2.RawDump)

    $("#dump3 #dumpRaw").text(obj.DataDump3.RawDump)

    $("#dump4 #dumpAuth").text(obj.DataDump4.Authorities)
    $("#dump4 #dumpIdent").text(obj.DataDump4.Identities)
    $("#dump4 #dumpMyNode").text(obj.DataDump4.MyNode)
  })
}

function updateTransactions() {
  resp = queryState("recentTransactions",function(resp){
    obj = JSON.parse(resp)
    //if($("#DBBlockHeight").text() != obj.DirectoryBlock.DBHeight) {
      $("#DBKeyMR > a").text(obj.DirectoryBlock.KeyMR)
      $("#DBBodyKeyMR").text(obj.DirectoryBlock.BodyKeyMR)
      $("#DBFullHash").text(obj.DirectoryBlock.FullHash)
      $("#DBBlockHeight").text(obj.DirectoryBlock.DBHeight)

      $("#panFactoids > #traxList > tbody").html("")
      if (obj.FactoidTransactions == null) {
        return
      }
      obj.FactoidTransactions.forEach(function(trans) {
        if(trans.TotalInput > 0.0001) {
          /*$("\
          <tr>\
              <td><a id='factom-search-link' type='facttransaction'>" + trans.TxID + "</a></td>\
              <td>" + trans.TotalInput + "</td>\
              <td>" + trans.TotalInputs + "</td>\
              <td>" + trans.TotalOutputs + "</td>\
          </tr>").insertBefore("#panFactoids > #traxList > tbody >tr:first")*/
          $("#panFactoids > #traxList > tbody").append("\
          <tr>\
              <td><a id='factom-search-link' type='factoidack'>" + trans.TxID + "</a></td>\
              <td>" + trans.TotalInput + "</td>\
              <td>" + trans.TotalInputs + "</td>\
              <td>" + trans.TotalOutputs + "</td>\
          </tr>")
        }
      })

      $("#panEntries > #traxList > tbody").html("")
      if(obj.Entries != null){
        obj.Entries.forEach(function(entry) {
          /*$("\
          <tr>\
              <td><a id='factom-search-link' type='entry'>" + entry.Hash + "</a></td>\
              <td><a id='factom-search-link' type='chainhead'>" + entry.ChainID  + "</a></td>\
              <td>" + entry.ContentLength + "</td>\
          </tr>").insertBefore("#panEntries > #traxList > tbody > tr:first")*/
          if (entry.ChainID == "Processing") {
            $("#panEntries > #traxList > tbody").append("\
            <tr>\
                <td><a id='factom-search-link' type='entry'>" + entry.Hash + "</a></td>\
                <td><a id='factom-search-link' type='chainhead'>" + entry.ChainID  + "</a></td>\
                <td>" + entry.ECCost + "</td>\
            </tr>")
          } else {
            $("#panEntries > #traxList > tbody").append("\
            <tr>\
                <td><a id='factom-search-link' type='entryack'>" + entry.Hash + "</a></td>\
                <td><a id='factom-search-link' type='chainhead'>" + entry.ChainID  + "</a></td>\
                <td>" + entry.ECCost + "</td>\
            </tr>")
          }
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
 //   }
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
    updateProgressBar("#syncSecond > .progress-meter", completeHeight, leaderHeight)
    percent = (completeHeight/leaderHeight) * 100
    percent = Math.floor(percent)
    $('#syncSecond > .progress-meter > .progress-meter-text').text(completeHeight + " of " + leaderHeight)
  })
}

function updateProgressBar(id, current, max) {
  percent = (current/max) * 100
  $(id).width(percent+ "%")
}

function updatePeerTotals() {
  resp = queryState("peerTotals", function(resp){
    if(resp.length == 0) {
      return
    }
    obj = JSON.parse(resp)
    if (typeof obj == "undefined") {
      $("#peerList > tfoot > tr > #peerquality").text("0")
    } else {
      $("#peerList > tfoot > tr > #peerquality").text(obj.PeerQualityAvg)
      $("#peerList > tfoot > tr > #up").text(formatBytes(obj.BytesSentTotal, obj.MessagesSent))
      $("#peerList > tfoot > tr > #down").text(formatBytes(obj.BytesReceivedTotal, obj.MessagesReceived))
    }
  })
}

function updatePeers() {
  resp = queryState("peers", function(resp){
    if(resp.length == 0) {
      return
    }
    obj = JSON.parse(resp)
    for (index in obj) {
      peer = obj[index]
      if($("#" + peer.Hash).length > 0) {
        con = peer.Connection
        if ($("#" + peer.Hash).find("#ip").val() != con.PeerAddress) {
          $("#" + peer.Hash).find("#ip span").text(con.PeerAddress)
          $("#" + peer.Hash).find("#ip span").attr("title", getIPCountry(con.PeerAddress))
          $("#" + peer.Hash).find("#ip").val(con.PeerAddress) // Value
          $("#" + peer.Hash).foundation()
        }
        if ($("#" + peer.Hash).find("#connected").val() != peer.Connected) {
          $("#" + peer.Hash).find("#connected").val(peer.Connected) // Value
          if(peer.Connected == false) { // Need to move to end
            $("#" + peer.Hash).delete("#peerList > tbody")
          }
          if (peer.Connected == true) {
            $("#" + peer.Hash).find("#connected").text("Connected")
          } else {
            $("#" + peer.Hash).find("#connected").text("Disconnected")
          }
        }
        if ($("#" + peer.Hash).find("#peerquality").val() != con.PeerQuality) {
          $("#" + peer.Hash).find("#peerquality").val(con.PeerQuality) // Value
          $("#" + peer.Hash).find("#peerquality").text(formatQuality(con.PeerQuality) + "/10")
        }
        if ($("#" + peer.Hash).find("#momentconnected").val() != con.MomentConnected) {
          $("#" + peer.Hash).find("#momentconnected").val(con.MomentConnected) // Value
          $("#" + peer.Hash).find("#momentconnected").text(peer.ConnectionTimeFormatted)
        }

        if ($("#" + peer.Hash).find("#sent").val().length == 0 || $("#" + peer.Hash).find("#sent").val() != con.BytesSent) {
          $("#" + peer.Hash).find("#sent").val(con.BytesSent) // Value
          $("#" + peer.Hash).find("#sent").text(formatBytes(con.BytesSent, con.MessagesSent))
        }
        if ($("#" + peer.Hash).find("#received").val().length == 0 || $("#" + peer.Hash).find("#received").val() != con.BytesReceived) {
          $("#" + peer.Hash).find("#received").val(con.BytesReceived) // Value
          $("#" + peer.Hash).find("#received").text(formatBytes(con.BytesReceived, con.MessagesReceived))
        }
      } else {
        // <td id='ip'><span data-tooltip class='has-tip top' title='ISP(geo130.comcast.net), Origin(USA)''>Loading...</span></td>\
        $("#peerList > tbody").prepend("\
        <tr id='" + peer.Hash + "'>\
            <td id='ip'><span data-tooltip class='has-tip top' title='ISP(???), Origin(???)'>Loading...</span></td>\
            <td id='connected'></td>\
            <td id='peerquality'></td>\
            <td id='momentconnected'></td>\
            <td id='sent' value='-10'></td>\
            <td id='received' value='-10'></td>\
            <td></td>\
        </tr>")

      }
    }
    updatePeerTotals()
  })
}

function getIPCountry(address){
  $.getJSON('http://ipinfo.io/' + address + '', function(data){
    console.log(data.country)
    return "ISP(" + data.org + ") Origin(" + data.country + ")"
  })
}

// 0-4  | -QR1 ...  -QR2
QUALITY_RANK_1 = -300
RANK_1_SCALE_MIN = 0
// 4-6  | -QR2 ... QR3
QUALITY_RANK_2 = -50
RANK_2_SCALE_MIN = 4
// 6-9 | QR3 ... QR4
QUALITY_RANK_3 = 100
RANK_3_SCALE_MIN = 6
// 9-10 | QR4 ... QR5
QUALITY_RANK_4 = 500
RANK_4_SCALE_MIN = 9
// 10   | QR5+
QUALITY_RANK_5 = 2000
RANK_5_SCALE_MIN = 10

function formatQuality(quality) {
  if (quality > QUALITY_RANK_5) { // QR4+
    return RANK_5_SCALE_MIN
  } else if (quality <= QUALITY_RANK_5 && quality >= QUALITY_RANK_4) { // QR3 ... QR4
    rankSpan = QUALITY_RANK_5 - QUALITY_RANK_4
    place = quality - QUALITY_RANK_4
    percent = place / rankSpan
    scaleSpan = RANK_5_SCALE_MIN - RANK_4_SCALE_MIN
    return Number(RANK_4_SCALE_MIN + percent * scaleSpan).toFixed(1)
  } else if (quality <= QUALITY_RANK_4 && quality >= QUALITY_RANK_3) { // QR3 ... QR4
    rankSpan = QUALITY_RANK_4 - QUALITY_RANK_3
    place = quality - QUALITY_RANK_3
    percent = place / rankSpan
    scaleSpan = RANK_4_SCALE_MIN - RANK_3_SCALE_MIN
    return Number(RANK_3_SCALE_MIN + percent * scaleSpan).toFixed(1)
  } else if (quality <= QUALITY_RANK_3 && quality >= QUALITY_RANK_2) { // QR2 ... QR3
    rankSpan = QUALITY_RANK_3 - QUALITY_RANK_2
    place = quality - QUALITY_RANK_2
    percent = place / rankSpan
    scaleSpan = RANK_3_SCALE_MIN - RANK_2_SCALE_MIN
    return Number(RANK_2_SCALE_MIN + percent * scaleSpan).toFixed(1)
  } else if (quality <= QUALITY_RANK_2 && quality >= QUALITY_RANK_1) { // QR1 ... QR2
    rankSpan = QUALITY_RANK_2 - QUALITY_RANK_1
    place = quality - QUALITY_RANK_1
    percent = place / rankSpan
    scaleSpan = RANK_2_SCALE_MIN - RANK_1_SCALE_MIN
    return Number(RANK_1_SCALE_MIN + percent * scaleSpan).toFixed(1)
  }  else { // QR0 -
    return 0
  }

}

function formatBytes(bytes, messages) {
  b = Number(bytes / 1000).toFixed(1)
  if (b < 100) {
    b = b + " KB"
  } if ((bytes / 1000000) < 100) {
    b = Number(bytes / 1000000).toFixed(1)
    b = b + " MB"
  } else {
      b = Number(bytes / 1000000000).toFixed(1)
    b = b + " GB"
  }
  m = messages
  if (m > 1000) {
    m = Number(messages / 1000).toFixed(1)
    m = m + " K"
  }
  return m + "(" + b + ")"
}