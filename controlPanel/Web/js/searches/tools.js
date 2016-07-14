$("#entry-external-id > #encoding").click(function() {
  encoding = jQuery(this).text()
  if(encoding.indexOf("Hex") > -1){
  	jQuery(this).find("a").text("Ascii: ")
  	str = convertFromHex(jQuery(this).parent().find("#data").text())
	jQuery(this).parent().find("#data").text(str)
  } else if (encoding.indexOf("Ascii") > -1){
	jQuery(this).find("a").text("Hex : ")
	str = convertToHex(jQuery(this).parent().find("#data").text())
	jQuery(this).parent().find("#data").text(str)
  }
})

function convertFromHex(hex) {
    var hex = hex.toString();//force conversion
    var str = '';
    for (var i = 0; i < hex.length; i += 2)
        str += String.fromCharCode(parseInt(hex.substr(i, 2), 16));
    return str;
}

function convertToHex(str) {
    var hex = '';
    for(var i=0;i<str.length;i++) {
        hex += ''+str.charCodeAt(i).toString(16);
    }
    return hex;
}

$("td > #entry-content-summary > a").click(function() {
  td = jQuery(this).parent().parent()
  td.find("#entry-content-summary").hide()
  td.find("#entry-content-body").show()//slideDown(100)
})

$("td > #entry-content-body > a").click(function() {
  td = jQuery(this).parent().parent()
    td.find("#entry-content-body").hide()
  td.find("#entry-content-summary").show()//slideDown(100)
})