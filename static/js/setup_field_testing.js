// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the Field Testing page.

var websocket;

// Sends a websocket message to play a given game sound on the audience display.
var playSound = function (sound) {
  websocket.send("playSound", sound);
};

// Sends a websocket message to set the hub light color.
var setHubLightColor = function (alliance) {
  var red = parseInt($("#" + alliance + "HubRed").val());
  var green = parseInt($("#" + alliance + "HubGreen").val());
  var blue = parseInt($("#" + alliance + "HubBlue").val());

  if (isNaN(red) || red < 0 || red > 255 ||
      isNaN(green) || green < 0 || green > 255 ||
      isNaN(blue) || blue < 0 || blue > 255) {
    alert("Please enter valid RGB values (0-255)");
    return;
  }

  websocket.send("setHubLightColor", {
    alliance: alliance,
    red: red,
    green: green,
    blue: blue
  });
};

// Sends a websocket message to set the hub active status.
var setHubActive = function (alliance, active) {
  websocket.send("setHubActive", {
    alliance: alliance,
    active: active
  });
};

// Handles a websocket message to update the PLC IO status.
var handlePlcIoChange = function (data) {
  $.each(data.Inputs, function (index, input) {
    $("#input" + index).text(input)
    $("#input" + index).attr("data-plc-value", input);
  });

  $.each(data.Registers, function (index, register) {
    $("#register" + index).text(register)
  });

  $.each(data.Coils, function (index, coil) {
    $("#coil" + index).text(coil)
    $("#coil" + index).attr("data-plc-value", coil);
  });
};

// Handles websocket messages for Red Hub PLC
var handleRedHubPlcIoChange = function (data) {
  $.each(data.Registers, function (index, register) {
    $("#redHubRegister" + index).text(register)
  });

  $.each(data.Coils, function (index, coil) {
    $("#redHubCoil" + index).text(coil)
    $("#redHubCoil" + index).attr("data-plc-value", coil);
  });
};

// Handles websocket messages for Blue Hub PLC
var handleBlueHubPlcIoChange = function (data) {
  $.each(data.Registers, function (index, register) {
    $("#blueHubRegister" + index).text(register)
  });

  $.each(data.Coils, function (index, coil) {
    $("#blueHubCoil" + index).text(coil)
    $("#blueHubCoil" + index).attr("data-plc-value", coil);
  });
};

$(function () {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/setup/field_testing/websocket", {
    plcIoChange: function (event) {
      handlePlcIoChange(event.data);
    },
    redHubPlcIoChange: function (event) {
      handleRedHubPlcIoChange(event.data);
    },
    blueHubPlcIoChange: function (event) {
      handleBlueHubPlcIoChange(event.data);
    }
  });
});
