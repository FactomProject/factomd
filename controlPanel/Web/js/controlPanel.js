var currentHeight = 0
var leaderHeight = 0

setInterval(updateHTML,1000);
setInterval(updateTransactions,1000);
setInterval(updateAllPeers,1000);
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
  updataDataDumps()
  //updatePeers()
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
  resp = queryState("dataDump", "",function(resp){
    obj = JSON.parse(resp)
    $("#dump1 #dumpShort").text(obj.DataDump1.ShortDump)
    $("#dump1 #dumpRaw").text(obj.DataDump1.RawDump)

    $("#dump2 #dumpRaw").text(obj.DataDump2.RawDump)

    $("#dump3 #dumpRaw").text(obj.DataDump3.RawDump)

    $("#dump4 #dumpAuth").text(obj.DataDump4.Authorities)
    $("#dump4 #dumpIdent").text(obj.DataDump4.Identities)
    $("#dump4 #dumpMyNode").text(obj.DataDump4.MyNode)

    $("#dump5 #dumpConRaw").text(obj.DataDump5.RawDump)
    $("#dump5 #dumpSort").text(obj.DataDump5.SortedDump)
  })
}

function updateTransactions() {
  resp = queryState("recentTransactions","",function(resp){
    obj = JSON.parse(resp)
    //if($("#DBBlockHeight").text() != obj.DirectoryBlock.DBHeight) {
      $("#DBKeyMR > a").text(obj.DirectoryBlock.KeyMR)
      $("#DBBodyKeyMR").text(obj.DirectoryBlock.BodyKeyMR)
      $("#DBFullHash").text(obj.DirectoryBlock.FullHash)
      $("#DBBlockHeight").text(obj.DirectoryBlock.DBHeight)
      $("#recent-directory-block").text(obj.DirectoryBlock.DBHeight)

      if(obj.FactoidTransactions != null){
        // Total
        $("#recent-factoid-total").text("(" + $("#panFactoids > #traxList > tbody > tr").length + ")")

        obj.FactoidTransactions.forEach(function(trans) {
          if(trans.TotalInput > 0.0001) {
            if($("#panFactoids > #traxList > tbody #" + trans.TxID).length > 0) {

            } else {
              $("#panFactoids > #traxList > tbody").prepend("\
              <tr id='" + trans.TxID + "'>\
                  <td><a id='factom-search-link' type='factoidack'>" + trans.TxID + "</a></td>\
                  <td>" + trans.TotalInput + "</td>\
                  <td>" + trans.TotalInputs + "</td>\
                  <td>" + trans.TotalOutputs + "</td>\
              </tr>")
            }
          }
        })
      }
      if(obj.Entries != null){
        obj.Entries.forEach(function(entry) {
          // Total
          $("#recent-entry-total").text("(" + $("#panEntries > #traxList > tbody > tr").length + ")")

          if ($("#panEntries > #traxList > tbody > tr").length > 100) {
            $("#panEntries > #traxList > tbody >tr").last().remove();
          }
          if ($("#panEntries #" + entry.Hash).length > 0) {
            if($("#"+entry.Hash + " #chainID a").text() != entry.ChainID) {
              $("#"+entry.Hash + " #chainID a").text(entry.ChainID)
            }
            if ($("#"+entry.Hash + " #chainID a").text() != "Processing") {
              $("#"+entry.Hash + " #entry-entryhash a").attr("type", "entry")
            }
            if($("#"+entry.Hash + " #eccost").text() != entry.ECCost) {
              $("#"+entry.Hash + " #eccost").text(entry.ECCost)
            }
          } else {
            $("#panEntries > #traxList > tbody").prepend("\
            <tr id='" + entry.Hash + "'>\
                <td id='entry-entryhash'><a id='factom-search-link' type='entryack'>" + entry.Hash + "</a></td>\
                <td id='chainID'><a id='factom-search-link' type='chainhead'>" + entry.ChainID  + "</a></td>\
                <td id='eccost'>" + entry.ECCost + "</td>\
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

// 3 Queriers in Batch
function getHeight() {
  resp = batchQueryState("myHeight,leaderHeight,completeHeight",function(resp){
    obj = JSON.parse(resp)
    respOne = obj[0].Height
    respTwo = obj[1].Height
    respThree = obj[2].Height

    currentHeight = parseInt(respOne)
    $("#nodeHeight").val(respOne)

    leaderHeight = parseInt(respTwo)
    updateProgressBar("#syncFirst > .progress-meter", currentHeight, leaderHeight)
    percent = (currentHeight/leaderHeight) * 100
    percent = Math.floor(percent)
    $('#syncFirst > .progress-meter > .progress-meter-text').text(percent + "% Synced (" + currentHeight + " of " + leaderHeight + ")")

    //$("#nodeHeight").val(resp)
    completeHeight = parseInt(respThree)
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

var peerHashes = [""]

// 2 Queries in Batch
function updateAllPeers() {
  batchQueryState("peerTotals,peers", function(respRaw){
    obj = JSON.parse(respRaw)
    respOne = obj[0]
    resp = obj[1]
    // Totals
    if(respOne.length == 0) {
      return
    }
    if (typeof obj == "undefined") {
      $("#peerList > tfoot > tr > #peerquality").text("0")
    } else {
      $("#peerList > tfoot > tr > #peerquality").text(formatQuality(obj.PeerQualityAvg) +"/10")
      $("#peerList > tfoot > tr > #up").text(formatBytes(obj.BytesSentTotal, obj.MessagesSent))
      $("#peerList > tfoot > tr > #down").text(formatBytes(obj.BytesReceivedTotal, obj.MessagesReceived))
    }
    // Table Body
    if(resp.length == 0) {
      return
    }
    for (index in resp) {
      peer = resp[index]
      peerHashes = [""]
      if($("#" + peer.Hash).length > 0) {
        peerHashes.push(peer.PeerHash)
        con = peer.Connection
        if ($("#" + peer.Hash).find("#ip").val() != peer.PeerHash) {
          $("#" + peer.Hash).find("#ip span").text(con.PeerAddress)
          //$("#" + peer.Hash).find("#ip span").attr("title", getIPCountry(con.PeerAddress))
          //$("#" + peer.Hash).find("#ip span").attr("title", con.ConnectionNotes)
          $("#" + peer.Hash).find("#ip").val(peer.PeerHash) // Value
          $("#" + peer.Hash).find("#disconnect").val(peer.PeerHash)

          // Reload Functions
          $("#" + peer.Hash).find("#disconnect").click(function(){
            queryState("disconnect",jQuery(this).val(), function(resp){
              console.log(resp)
            })
          })
          $("#" + peer.Hash).foundation()
        }
        if ($("#" + peer.Hash).find("#ip span").attr("title") != con.ConnectionNotes) {
          element = $("#" + peer.Hash).find("#ip span") 
          wich = $("has-tip").index(element); 
          $(".tooltip").eq(wich).html(con.ConnectionNotes); 

        }
        if ($("#" + peer.Hash).find("#connected").val() != con.ConnectionState) {
          $("#" + peer.Hash).find("#connected").val(con.ConnectionState) // Value
          $("#" + peer.Hash).find("#connected").text(con.ConnectionState)

          if(peer.Connected == false) { // Need to move to end
            $("#peerList > tbody").find(("#" + peer.Hash)).remove()
          }
          /*if (peer.Connected == true) {
            $("#" + peer.Hash).find("#connected").text("Connected")
          } else {
            $("#" + peer.Hash).find("#connected").text("Disconnected")
          }*/
        }
        if ($("#" + peer.Hash).find("#peerquality").val() != con.PeerQuality) {
          $("#" + peer.Hash).find("#peerquality").val(con.PeerQuality) // Value
          $("#" + peer.Hash).find("#peerquality").text(formatQuality(con.PeerQuality) + "/10")
        }

        if ($("#" + peer.Hash).find("#sent").val().length == 0 || $("#" + peer.Hash).find("#sent").val() != con.BytesSent) {
          $("#" + peer.Hash).find("#sent").val(con.BytesSent) // Value
          $("#" + peer.Hash).find("#sent").text(formatBytes(con.BytesSent, con.MessagesSent))
        }
        if ($("#" + peer.Hash).find("#received").val().length == 0 || $("#" + peer.Hash).find("#received").val() != con.BytesReceived) {
          $("#" + peer.Hash).find("#received").val(con.BytesReceived) // Value
          $("#" + peer.Hash).find("#received").text(formatBytes(con.BytesReceived, con.MessagesReceived))
        }
        if ($("#" + peer.Hash).find("#momentconnected").val() != peer.ConnectionTimeFormatted) {
          $("#" + peer.Hash).find("#momentconnected").val(peer.ConnectionTimeFormatted) // Value
          $("#" + peer.Hash).find("#momentconnected").text(peer.ConnectionTimeFormatted)
        }
      } else {
        // <td id='ip'><span data-tooltip class='has-tip top' title='ISP(geo130.comcast.net), Origin(USA)''>Loading...</span></td>\
        $("#peerList > tbody").prepend("\
        <tr id='" + peer.Hash + "'>\
            <td id='ip'><span data-tooltip class='has-tip top' title=''>Loading...</span></td>\
            <td id='connected'></td>\
            <td id='peerquality'></td>\
            <td id='momentconnected'></td>\
            <td id='sent' value='-10'></td>\
            <td id='received' value='-10'></td>\
            <td><a id='disconnect' class='button tiny alert'>Disconnect</a></td>\
        </tr>")

      }
    }
  }) 
}

function getIPCountry(address){
 /* $.getJSON('http://ipinfo.io/' + address + '', function(data){
    console.log(data.country)
    return "ISP(" + data.org + ") Origin(" + data.country + ")"
  })*/
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
  } else if ((bytes / 1000000) < 100) {
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