var currentHeight = 0
var leaderHeight = 0

setInterval(updateHTML,3000);
var serverOnline = false
// Used to update some things less frequently
var skipInterval = false

$(window).load(
    function() {
      updateHTML()
      setTimeout(function () {
            updateHTML()
      }, 1000);
    }
);

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

  if ($("#indexnav-main").hasClass("is-active")) {
    // Main Tab
    updateHeight() 
    updateAllPeers()
    // Does every another cycle
    if(!skipInterval){
      updateTransactions()
      skipInterval = true
    } else {
      skipInterval = false
    }
  } else if($("#indexnav-more").hasClass("is-active")) {
    // Detailed Tab
    updataDataDumps()
  }

}

// Update when we switch tabs
$(".tabs-control-panel li a").click(function(){
  setTimeout(
  function() 
  {
    updateHTML()
  }, 300)
})

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

      // Total
      $("#recent-factoid-total").text("(" + $("#panFactoids > #traxList > tbody > tr").length + ")")

      $("section #factom-search-link").on('click',function(e) {
        type = jQuery(this).attr("type")
        hash = jQuery(this).text()
        if (hash == "Processing") {
          return
        }
        var x = new XMLHttpRequest()
        x.onreadystatechange = function() {
          if(x.readyState == 4) {
            if(e.which == 1){
              window.location = "search?input=" + hash + "&type=" + type
            } else if(e.which == 2) {
              window.open("/search?input=" + hash + "&type=" + type);
            }
          }
        }
        var formDataLink = new FormData();
        formDataLink.append("method", "search")
        formDataLink.append("search", hash)

        x.open("POST", "/post")
        x.send(formDataLink)
      })
  })
}

// 3 Queriers in Batch
function updateHeight() {
  resp = batchQueryState("myHeight,leaderHeight,completeHeight,servercount,channelLength",function(resp){
    obj = JSON.parse(resp)
    myHeight = obj[0].Height
    lHeight = obj[1].Height
    compHeight = obj[2].Height
    feds = obj[3].fed
    auds = obj[3].aud
    respFive = obj[4].length

    $("#serverfedcount").val(feds)
    $("#serveraudcount").val(auds)

    currentHeight = parseInt(myHeight)
    $("#nodeHeight").val(myHeight)

    leaderHeight = parseInt(lHeight)
    updateProgressBar("#syncFirst > .progress-meter", currentHeight, leaderHeight)
    percent = 0
    if(leaderHeight == 0) {
      percent = 100
    } else {
      percent = (currentHeight/leaderHeight) * 100
      percent = Math.floor(percent)
    }
    $('#syncFirst > .progress-meter > .progress-meter-text').text(percent + "% Synced (" + currentHeight + " of " + leaderHeight + ")")

    //$("#nodeHeight").val(resp)
    completeHeight = parseInt(compHeight)
    updateProgressBar("#syncSecond > .progress-meter", completeHeight, leaderHeight)
    percent = (completeHeight/leaderHeight) * 100
    percent = Math.floor(percent)
    $('#syncSecond > .progress-meter > .progress-meter-text').text(completeHeight + " of " + leaderHeight)

    // DisplayState Channel length
    // console.log("Chan Length:", respFive)
  })
}

function updateProgressBar(id, current, max) {
  if(max == 0) {
    percent = (current/max) * 100
    $(id).width("100%")
  } else {
    percent = (current/max) * 100
    $(id).width(percent+ "%")
  }
}

var peerHashes = [""]

// 2 Queries in Batch
function updateAllPeers() {
  batchQueryState("peerTotals,peers", function(respRaw){
    obj = JSON.parse(respRaw)
    respOne = obj[0]
    resp = obj[1]
    $("#totalPeerCount").text(resp.length)

    // Totals
    if(respOne.length == 0) {
      return
    }
    if (typeof respOne == "undefined") {
      //$("#peerList > tfoot > tr > #peerquality").text("0")
    } else {
      //$("#peerList > tfoot > tr > #peerquality").text(formatQuality(obj.PeerQualityAvg))
      $("#peerList > tfoot > tr > #up").text(formatBytes(respOne.BytesSentTotal, respOne.MessagesSent))
      $("#peerList > tfoot > tr > #down").text(formatBytes(respOne.BytesReceivedTotal, respOne.MessagesReceived))
    }
    // Table Body
    if(resp.length == 0) {
        $("#peerList tbody tr").each(function(){
          jQuery(this).remove()
        })
      return
    }
    peerHashes = [""]

    // To avoid hundreds of new html elements updated in a quick span, it will be limited.
    newPeers = 0
    for (index in resp) {
      peer = resp[index]
      peerHashes.push(peer.PeerHash)
      if($("#" + peer.Hash).length > 0) {
        con = peer.Connection
        if ($("#" + peer.Hash).find("#ip").val() != peer.PeerHash) {
          $("#" + peer.Hash).find("#ip span").text(con.PeerAddress)
          $("#" + peer.Hash).find("#ip").val(peer.PeerHash) // Value
          $("#" + peer.Hash).find("#disconnect").attr("value", peer.PeerHash)

          $("#" + peer.Hash).find("#disconnect").click(function(){
            queryState("disconnect", jQuery(this).attr("value"), function(resp){
              obj = JSON.parse(resp)
              console.log(obj)
              if(obj.Access == "denied") {
                $("#" + obj.Id).find("#disconnect").addClass("disabled")
                $("#" + obj.Id).find("#disconnect").text("Denied")
              } else {
                  $("#" + obj.Id).find("#disconnect").addClass("disabled")
                  $("#" + obj.Id).find("#disconnect").text("Attempting")
              }
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
          if(!($("#" + peer.Hash).hasClass(formatQuality(con.PeerQuality)))){
            $("#" + peer.Hash).removeClass()
            $("#" + peer.Hash).addClass(formatQuality(con.PeerQuality))
          }
          //$("#" + peer.Hash).find("#peerquality").text(formatQuality(con.PeerQuality))
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
        newPeers = newPeers + 1
        if (newPeers < 20) { // If over 20 new peers, only load 20. Will get remaining next pass.
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
    }
    // Cleanup Routine
    $("#peerList tbody tr").each(function(){
      if(!jQuery(this).find("#ip span").text().includes("Loading")){
        if(!contains(peerHashes, jQuery(this).find("#ip").val())) {
         jQuery(this).remove()
        }
      } else {
        if(jQuery(this).find("#ip span").text() == "Loading...."){
          jQuery(this).remove()
        } else {
          jQuery(this).find("#ip span").text("Loading....")
        }
      }
    })
  }) 
}

function contains(haystack, needle) {
    var i = haystack.length;
    while (i--) {
       if (haystack[i] === needle) {
           return true;
       }
    }
    return false;
}


function getIPCountry(address){
 /* $.getJSON('http://ipinfo.io/' + address + '', function(data){
    console.log(data.country)
    return "ISP(" + data.org + ") Origin(" + data.country + ")"
  })*/
}

// Using two logistic functions
function formatQuality(quality) {
  if(quality >= 100) {
    return "rank-green"
  } else if(quality >= -50) {
    return "rank-gold"
  } else {
    return "rank-red"
  }
  /*quality = quality + 300
  if(quality < 0) {
    return 0
  } else if(quality > 3000) {
    return 10
  } else if(quality < 390) {
    limit = 8
    exponent = (-.5) * ((quality * .02) - 5)
    q = limit / (1+ Math.pow(Math.E,exponent))
    return Number(q).toFixed(1)
  } else {
    limit = 4
    exponent = (-.3) * ((quality - 60) * 0.008 - 5)
    q = limit / (1 + (Math.pow(Math.E,exponent))) + 6
    return Number(q).toFixed(1)
  }*/
}

function formatBytes(bytes, messages) {
  if (bytes == undefined || messages == undefined) {
    return "0 (0 Kb)"
  }
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