var currentHeight = 0
var leaderHeight = 0
var SortingFunction = SortDuration;

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
  $.ajax('./', {
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
    $("#dump1 #dumpSyncing").text(obj.DataDump1.SyncingDump)

    $("#dump2 #dumpRawProc").text(obj.DataDump2.RawDump)
    $("#dump2 #dumpNext").text(obj.DataDump2.NextDump)
    $("#dump2 #dumpPrev").text(obj.DataDump2.PrevDump)


    $("#dump3 #dumpRaw").text(obj.DataDump3.RawDump)

    $("#dump4 #dumpAuth").text(obj.DataDump4.Authorities)
    $("#dump4 #dumpIdent").text(obj.DataDump4.Identities)
    $("#dump4 #dumpMyNode").text(obj.DataDump4.MyNode)

    $("#dump5 #dumpConRaw").text(obj.DataDump5.RawDump)
    $("#dump5 #dumpSort").text(obj.DataDump5.SortedDump)

    $("#dump6 #dumpElections").text(obj.ElectionDataDump.Elections)
    $("#dump6 #dumpSimulatedElections").text(obj.ElectionDataDump.SimulatedElection)
    if(obj.LogSettingsDump.CurrentLogSettings == "") {
        obj.LogSettingsDump.CurrentLogSettings = " "
    }
    $("#dump7 #current-log-value").text(obj.LogSettingsDump.CurrentLogSettings)
  })
}

function updateTransactions() {
  resp = queryState("recentTransactions","",function(resp){
    obj = JSON.parse(resp)
    //if($("#DBBlockHeight").text() != obj.DirectoryBlock.DBHeight) {
      $("#DBKeyMR > a").text(obj.DirectoryBlock.KeyMR)
      $("#DBBodyKeyMR").text(obj.DirectoryBlock.BodyKeyMR)
      $("#DBFullHash").text(obj.DirectoryBlock.FullHash)
      $("#DBBlockTimestamp").text(obj.DirectoryBlock.Timestamp)
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
              if ($("#panFactoids > #traxList > tbody > tr").length > 100) {
                $("#panFactoids > #traxList > tbody >tr").last().remove();
              } 
            }
          }
        })
      }
      if(obj.Entries != null){
        obj.Entries.forEach(function(entry) {
          // Total
          $("#recent-entry-total").text("(" + $("#panEntries > #traxList > tbody > tr").length + ")")
          if ($("#panEntries #" + entry.Hash).length > 0) {
            if($("#"+entry.Hash + " #chainID a").text() != entry.chainid) {
              $("#"+entry.Hash + " #chainID a").text(entry.chainid)
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
                <td id='chainID'><a id='factom-search-link' type='chainhead'>" + entry.chainid  + "</a></td>\
                <td id='eccost'>" + entry.ECCost + "</td>\
            </tr>")
            if ($("#panEntries > #traxList > tbody > tr").length > 100) {
              $("#panEntries > #traxList > tbody >tr").last().remove();
            }
          }
        })
      }

      // Total
      $("#recent-factoid-total").text("(" + $("#panFactoids > #traxList > tbody > tr").length + ")")
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
    percentSecond = 0
    if(leaderHeight == 0) {
      percent = 100
    } else {
      percentSecond = (completeHeight/leaderHeight) * 100
      percentSecond = Math.floor(percentSecond)
    }
    $('#syncSecond > .progress-meter > .progress-meter-text').text(percentSecond + "% Synced (" + completeHeight + " of " + leaderHeight +")")

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

function updatePeerIcon(element, connection) {
    switch (connection.PeerType) {
        case "special_config":
            element.removeClass("hidden");
            element.prop("title", "Special Peer (configuration file)");
            break;
        case "special_cmdline":
            element.removeClass("hidden");
            element.prop("title", "Special Peer (command line)");
            break;
        case "regular":
        default:
            element.addClass("hidden");
            element.prop("title", "");
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
      $("#peerList > tfoot > tr > #data-up").text(formatBytes(respOne.BytesSentTotal))
      $("#peerList > tfoot > tr > #data-down").text(formatBytes(respOne.BytesReceivedTotal))
      $("#peerList > tfoot > tr > #speed-up").text(formatBps(respOne.BPSUp))
      $("#peerList > tfoot > tr > #speed-down").text(formatBps(respOne.BPSDown))
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
        con = peer.Connection;
        if ($("#" + peer.Hash).find("#ip").val() != peer.PeerHash) {
          $("#" + peer.Hash).find("#ip span").text(con.PeerAddress)
          $("#" + peer.Hash).find("#ip").val(peer.PeerHash) // Value
          $("#" + peer.Hash).foundation()
        }
        if ($("#" + peer.Hash).find("#ip span").attr("title") != con.ConnectionNotes) {
          element = $("#" + peer.Hash).find("#ip span") 
          wich = $("has-tip").index(element); 
          $(".tooltip").eq(wich).html(con.Hash);

        }
        if ($("#" + peer.Hash).find("#protocol").val() != con.ConnectionState) {
          $("#" + peer.Hash).find("#protocol").val(con.ConnectionState) // Value
          $("#" + peer.Hash).find("#protocol").text(con.ConnectionState)
        }
        if(peer.Connected == false) { // Need to move to end
            $("#peerList > tbody").find(("#" + peer.Hash)).remove()
        }

        if ($("#" + peer.Hash).find("#up").val().length == 0 || $("#" + peer.Hash).find("#up").val() != con.BPSUp) {
            $("#" + peer.Hash).find("#up").val(con.BPSUp) // Value
            $("#" + peer.Hash).find("#up").text(formatBps(con.BPSUp))
        }
        if ($("#" + peer.Hash).find("#down").val().length == 0 || $("#" + peer.Hash).find("#down").val() != con.BPSDown) {
            $("#" + peer.Hash).find("#down").val(con.BPSDown) // Value
            $("#" + peer.Hash).find("#down").text(formatBps(con.BPSDown))
        }
        if ($("#" + peer.Hash).find("#sent").val().length == 0 || $("#" + peer.Hash).find("#sent").val() != con.BytesSent) {
          $("#" + peer.Hash).find("#sent").val(con.BytesSent) // Value
          $("#" + peer.Hash).find("#sent").text(formatBytes(con.BytesSent))
        }
        if ($("#" + peer.Hash).find("#received").val().length == 0 || $("#" + peer.Hash).find("#received").val() != con.BytesReceived) {
          $("#" + peer.Hash).find("#received").val(con.BytesReceived) // Value
          $("#" + peer.Hash).find("#received").text(formatBytes(con.BytesReceived))
          if(!($("#" + peer.Hash).hasClass(formatQuality(con.BytesReceived)))){
            $("#" + peer.Hash).removeClass()
            $("#" + peer.Hash).addClass(formatQuality(con.BytesReceived))
          }
        }
        if ($("#" + peer.Hash).find("#momentconnected").val() != peer.ConnectionTimeFormatted) {
          $("#" + peer.Hash).find("#momentconnected").val(peer.ConnectionTimeFormatted) // Value
          $("#" + peer.Hash).find("#momentconnected").text(peer.ConnectionTimeFormatted)
        }
      } else {
        newPeers = newPeers + 1
        if (newPeers < 20) { // If over 20 new peers, only load 20. Will get remaining next pass.
            if(PeerAddFromTopToggle == false) {
              $("#peerList > tbody").prepend("\
              <tr id='" + peer.Hash + "'>\
                  <td id='ip'>\
                      <i class='fa fa-link hidden'></i>\
                      <span data-tooltip class='has-tip top' title=''>Loading...</span></td>\
                  <td id='protocol'></td>\
                  <td id='momentconnected'></td>\
                  <td id='down' value='-10'></td>\
                  <td id='up' value='-10'></td>\
                  <td id='received' value='-10'></td>\
                  <td id='sent' value='-10'></td>\
              </tr>")
            } else {
              $("#peerList > tbody").append("\
              <tr id='" + peer.Hash + "'>\
                  <td id='ip'>\
                      <i class='fa fa-link hidden'></i>\
                      <span data-tooltip class='has-tip top' title=''>Loading...</span></td>\
                  <td id='protocol'></td>\
                  <td id='momentconnected'></td>\
                  <td id='down' value='-10'></td>\
                  <td id='up' value='-10'></td>\
                  <td id='received' value='-10'></td>\
                  <td id='sent' value='-10'></td>\
              </tr>")
            }
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

    SortingFunction();
  }) 
}

// Add listeners to disconnect buttons
$("body").on('mouseup',"#peerList  #disconnect",function(e) {
  queryState("disconnect", jQuery(this).attr("value"), function(resp){
    obj = JSON.parse(resp)
    if(obj.Access == "denied") {
      $("#" + obj.Id).find("#disconnect").addClass("disabled")
      $("#" + obj.Id).find("#disconnect").text("Denied")
    } else {
        $("#" + obj.Id).find("#disconnect").addClass("disabled")
        $("#" + obj.Id).find("#disconnect").text("Attempting")
    }
  })
})


SortToggle = true
PeerAddFromTopToggle = true
// Sorting
// Sort by Duration
function SortDuration () {
    SortingFunction = SortDuration;
    $(".sorting-img").addClass("hide")
    if(SortToggle == true) {
      $("#peer-duration-sort-img").removeClass("hide")
      $("#peer-duration-sort-img").attr("src","img/up.png")
    } else {
      $("#peer-duration-sort-img").removeClass("hide")
      $("#peer-duration-sort-img").attr("src","img/down.png")
    }
  
    array = $("#peerList tbody tr").get()
    valArray = $("#peerList tbody tr").find("#momentconnected").get()
  
    array = generalSort(durationIsLessThan, array, valArray)
  
    $("#peerList tbody").html(array)
  
    PeerAddFromTopToggle = SortToggle
}

function SortIP() {
    SortingFunction = SortIP;
    $(".sorting-img").addClass("hide")
    if(SortToggle == true) {
      $("#peer-ip-sort-img").removeClass("hide")
      $("#peer-ip-sort-img").attr("src","img/up.png")
    } else {
      $("#peer-ip-sort-img").removeClass("hide")
      $("#peer-ip-sort-img").attr("src","img/down.png")
    }
  
    array = $("#peerList tbody tr").get()
    valArray = $("#peerList tbody tr").find("#ip span").get()
  
    array = generalSort(ipIsLessThan, array, valArray)
  
    $("#peerList tbody").html(array)
}

function SortDownrate() {
    SortingFunction = SortDownrate;
    $(".sorting-img").addClass("hide")
    if(SortToggle == true) {
      $("#peer-down-sort-img").removeClass("hide")
      $("#peer-down-sort-img").attr("src","img/up.png")
    } else {
      $("#peer-down-sort-img").removeClass("hide")
      $("#peer-down-sort-img").attr("src","img/down.png")
    }
  
    array = $("#peerList tbody tr").get()
    valArray = $("#peerList tbody tr").find("#down").get()
  
    array = generalSort(msgIsLessThan, array, valArray)
  
    $("#peerList tbody").html(array)
    PeerAddFromTopToggle = SortToggle
}



$("#peer-duration").on('mouseup', function(e){
    SortToggle = !SortToggle;
    SortDuration();
    
})

// Sort by IP
$("#peer-ip").on('mouseup', function(e){
    SortToggle = !SortToggle;
    SortIP();
    
})

// Sort by Downrate
$("#peer-down").on('mouseup', function(e){
    SortToggle = !SortToggle;
    SortDownrate();
})


function SortUprate() {
    SortingFunction = SortUprate;
    $(".sorting-img").addClass("hide")
    if(SortToggle == true) {
      $("#peer-up-sort-img").removeClass("hide")
      $("#peer-up-sort-img").attr("src","img/up.png")
    } else {
      $("#peer-up-sort-img").removeClass("hide")
      $("#peer-up-sort-img").attr("src","img/down.png")
    }
  
    array = $("#peerList tbody tr").get()
    valArray = $("#peerList tbody tr").find("#up").get()
  
    array = generalSort(msgIsLessThan, array, valArray)
  
    $("#peerList tbody").html(array)
    PeerAddFromTopToggle = SortToggle
}
// Sort by Downrate
$("#peer-up").on('mouseup', function(e){
    SortToggle = !SortToggle;
    SortUprate();
})

function SortSent() {
    SortingFunction = SortSent;
    $(".sorting-img").addClass("hide")
    if(SortToggle == true) {
      $("#peer-sent-sort-img").removeClass("hide")
      $("#peer-sent-sort-img").attr("src","img/up.png")
    } else {
      $("#peer-sent-sort-img").removeClass("hide")
      $("#peer-sent-sort-img").attr("src","img/down.png")
    }
  
    array = $("#peerList tbody tr").get()
    valArray = $("#peerList tbody tr").find("#sent").get()
  
    array = generalSort(msgIsLessThan, array, valArray)
  
    $("#peerList tbody").html(array)
    PeerAddFromTopToggle = SortToggle
}

// Sort by Sent
$("#peer-sent").on('mouseup', function(e){
    SortToggle = !SortToggle;
    SortSent();
})


function SortReceived() {
    SortingFunction = SortReceived;
    $(".sorting-img").addClass("hide")
    if(SortToggle == true) {
      $("#peer-received-sort-img").removeClass("hide")
      $("#peer-received-sort-img").attr("src","img/up.png")
    } else {
      $("#peer-received-sort-img").removeClass("hide")
      $("#peer-received-sort-img").attr("src","img/down.png")
    }
  
    array = $("#peerList tbody tr").get()
    valArray = $("#peerList tbody tr").find("#received").get()
  
    array = generalSort(msgIsLessThan, array, valArray)
  
    $("#peerList tbody").html(array)
    PeerAddFromTopToggle = SortToggle
}
// Sort by Received
$("#peer-received").on('mouseup', function(e){
    SortToggle = !SortToggle;
    SortReceived();
})

function generalSort(lessThanFunction, array, valueArray) {
  peerLen = valueArray.length
  for(index = 0; index < peerLen; index++) {
    tmpVal = valueArray[index]
    tmp = array[index]

    if(SortToggle == true) {
      for(j = index - 1; j > -1 && !lessThanFunction(valueArray[j].innerText, tmpVal.innerText); j--) {
        valueArray[j+1] = valueArray[j]
        array[j+1] = array[j]
      }
    } else {
      for(j = index - 1; j > -1 && lessThanFunction(valueArray[j].innerText, tmpVal.innerText); j--) {
        valueArray[j+1] = valueArray[j]
        array[j+1] = array[j]
      }
    }

    valueArray[j+1] = tmpVal
    array[j+1] = tmp
  }
  return array
}

function msgIsLessThan(a, b) {
  if(typeof a != "string" || typeof b != "string") {
    return 0
  }
  if(a.length == 0 || b.length == 0) {
    return 0
  }

  aSplit = a.split(" ")
  aVal = convertToBytes(aSplit[0], aSplit[1])

  bSplit = b.split(" ")
  bVal = convertToBytes(bSplit[0], bSplit[1])

  if(aVal < bVal) {
    return 1
  }
  return 0

}

function ipIsLessThan(a, b) {
  if(typeof a != "string" || typeof b != "string") {
    return -1
  }
  a.split(".")
  b.split(".")

  aLen = a.length
  bLen = b.length
  if(aLen < bLen) {
    return 1
  }

  for(i = 0; i < aLen; i++){
    if(Number(b[i]) == "NaN") {
      return 1
    } else if(Number(a[i]) == "NaN"){
      return 0
    }

    if(Number(a[i]) < Number(b[i])) {
      return 1
    } else if(Number(a[i]) > Number(b[i])){
      return 0
    }
  }
}

function durationIsLessThan(a, b) {
  aSec = convertToSeconds(a)
  bSec = convertToSeconds(b)
 // console.log(a, aSec,"|",b, bSec)
  if(aSec == -1 || bSec == -1) {
    return -1
  }
  if(aSec <= bSec) {
    return 1 // True
  } else {
    return 0 // False
  }
}

function convertToSeconds(time) {
  if(typeof time != "string") {
    return -1
  }
  var seconds = time.split(" ");
  if(seconds.length < 2) {
    return -1
  }

  // If there is a 0 min/hr/day, it should still greater
  // than lower denomination. Adding 1 covers the 0 case
  seconds[0]++
  if(seconds[1].includes("sec")) {
    return seconds[0] * 1
  } else if(seconds[1].includes("min")) {
    return seconds[0] * 60
  } else if(seconds[1].includes("hr")) {
    return seconds[0] * 3600
  } else if(seconds[1].includes("day")) {
    return seconds[0] * 86400
  }
}

function convertToBytes(number, string) {
  if(string.includes("KB") || string.includes("Kbps")) {
    return number * 1e+3;
  } else if(string.includes("MB") || string.includes("Mbps")) {
    return number * 1e+6;
  } else if(string.includes("GB") || string.includes("Gbps")) {
    return number * 1e+9;
  }
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
  if(quality >= 10485760) { // 10 MB
    return "rank-green"
  } else if(quality >= 1048576) { // 1 MB
    return "rank-gold"
  } else {
    return "rank-red"
  }
}

function formatBytes(bytes) {
    if (bytes == undefined) {
      return "0 Kb"
    }
    b = Number(bytes / 1e+3).toFixed(1)
    if (b < 100) {
      b = b + " KB"
    } else if ((bytes / 1e+6) < 100) {
      b = Number(bytes / 1e+6).toFixed(1)
      b = b + " MB"
    } else {
        b = Number(bytes / 1e+9).toFixed(1)
      b = b + " GB"
    }
    return b;
}
function formatBps(bytes) {
    if (bytes == undefined) {
      return "0 Kbps"
    }

    var bits = bytes * 8;

    if (bits > 1e+9) {
        return Number(bits / 1e+9).toFixed(1) + " Gbps";
    }
    if (bits > 1e+6) {
        return Number(bits / 1e+6).toFixed(1) + " Mbps";
    }
    return Number(bits / 1e+3).toFixed(1) + " Kbps";
}
 

function formatBytesMessages(bytes, messages) {
  if (bytes == undefined || messages == undefined) {
    return "0 (0 Kb)"
  }
  b = Number(bytes / 1e+3).toFixed(1)
  if (b < 100) {
    b = b + " KB"
  } else if ((bytes / 1e+6) < 100) {
    b = Number(bytes / 1e+6).toFixed(1)
    b = b + " MB"
  } else {
      b = Number(bytes / 1e+9).toFixed(1)
    b = b + " GB"
  }
  m = messages
  if (m > 1000) {
    m = Number(messages / 1000).toFixed(1)
    m = m + " K"
  }
  return m + "(" + b + ")"
}
