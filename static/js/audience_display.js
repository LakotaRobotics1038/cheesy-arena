// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Author: nick@team254.com (Nick Eyre)
//
// Client-side methods for the audience display.

var websocket;
let transitionMap;
const transitionQueue = [];
let transitionInProgress = false;
let currentScreen = "blank";
let redSide;
let blueSide;
let currentMatch;
let overlayCenteringHideParams;
let overlayCenteringShowParams;
const allianceSelectionTemplate = Handlebars.compile($("#allianceSelectionTemplate").html());
const sponsorImageTemplate = Handlebars.compile($("#sponsorImageTemplate").html());
const sponsorTextTemplate = Handlebars.compile($("#sponsorTextTemplate").html());

// Constants for overlay positioning. The CSS is the source of truth for the values that represent initial state.
const overlayCenteringTopUp = "-130px";
const overlayCenteringBottomHideParams = {queue: false, bottom: $("#overlayCentering").css("bottom")};
const overlayCenteringBottomShowParams = {queue: false, bottom: "0px"};
const overlayCenteringTopHideParams = {queue: false, top: overlayCenteringTopUp};
const overlayCenteringTopShowParams = {queue: false, top: "50px"};
const eventMatchInfoDown = "30px";
const eventMatchInfoUp = $("#eventMatchInfo").css("height");
const logoUp = "20px";
const logoDown = $("#logo").css("top");
const scoreIn = $(".score").css("width");
const scoreMid = "185px";
const scoreOut = "370px";
const scoreFieldsOut = "150px";
const scoreLogoTop = "-530px";
const bracketLogoTop = "-780px";
const bracketLogoScale = 0.75;
const timeoutDetailsIn = $("#timeoutDetails").css("width");
const timeoutDetailsOut = "570px";

// Handles a websocket message to change which screen is displayed.
const handleAudienceDisplayMode = function (targetScreen) {
  transitionQueue.push(targetScreen);
  executeTransitionQueue();
};

// Sequentially executes all transitions in the queue. Returns without doing anything if another invocation is already
// in progress.
const executeTransitionQueue = function () {
  if (transitionInProgress) {
    // There is an existing invocation of this method which will execute all transitions in the queue.
    return;
  }

  if (transitionQueue.length > 0) {
    transitionInProgress = true;
    const targetScreen = transitionQueue.shift();
    const callback = function () {
      // When the current transition is complete, call this method again to invoke the next one in the queue.
      currentScreen = targetScreen;
      transitionInProgress = false;
      setTimeout(executeTransitionQueue, 100);  // A small delay is needed to avoid visual glitches.
    };

    if (targetScreen === currentScreen) {
      callback();
      return;
    }

    if (targetScreen === "sponsor") {
      initializeSponsorDisplay();
    }

    let transitions = transitionMap[currentScreen][targetScreen];
    if (transitions !== undefined) {
      transitions(callback);
    } else {
      // There is no direct transition defined; need to go to the blank screen first.
      transitionMap[currentScreen]["blank"](function () {
        transitionMap["blank"][targetScreen](callback);
      });
    }
  }
};

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function (data) {
  currentMatch = data.Match;
  $(`#${redSide}Team1`).text(currentMatch.Red1);
  $(`#${redSide}Team1`).parent().attr("data-yellow-card", data.Teams["R1"]?.YellowCard);
  $(`#${redSide}Team2`).text(currentMatch.Red2);
  $(`#${redSide}Team2`).parent().attr("data-yellow-card", data.Teams["R2"]?.YellowCard);
  $(`#${redSide}Team3`).text(currentMatch.Red3);
  $(`#${redSide}Team3`).parent().attr("data-yellow-card", data.Teams["R3"]?.YellowCard);
  $(`#${redSide}Team1Avatar`).attr("src", getAvatarUrl(currentMatch.Red1));
  $(`#${redSide}Team2Avatar`).attr("src", getAvatarUrl(currentMatch.Red2));
  $(`#${redSide}Team3Avatar`).attr("src", getAvatarUrl(currentMatch.Red3));
  $(`#${blueSide}Team1`).text(currentMatch.Blue1);
  $(`#${blueSide}Team1`).parent().attr("data-yellow-card", data.Teams["B1"]?.YellowCard);
  $(`#${blueSide}Team2`).text(currentMatch.Blue2);
  $(`#${blueSide}Team2`).parent().attr("data-yellow-card", data.Teams["B2"]?.YellowCard);
  $(`#${blueSide}Team3`).text(currentMatch.Blue3);
  $(`#${blueSide}Team3`).parent().attr("data-yellow-card", data.Teams["B3"]?.YellowCard);
  $(`#${blueSide}Team1Avatar`).attr("src", getAvatarUrl(currentMatch.Blue1));
  $(`#${blueSide}Team2Avatar`).attr("src", getAvatarUrl(currentMatch.Blue2));
  $(`#${blueSide}Team3Avatar`).attr("src", getAvatarUrl(currentMatch.Blue3));

  // Show alliance numbers if this is a playoff match.
  if (currentMatch.Type === matchTypePlayoff) {
    $(`#${redSide}PlayoffAlliance`).text(currentMatch.PlayoffRedAlliance);
    $(`#${blueSide}PlayoffAlliance`).text(currentMatch.PlayoffBlueAlliance);
    $(".playoff-alliance").show();

    // Show the series status if this playoff round isn't just a single match.
    if (data.Matchup.NumWinsToAdvance > 1) {
      $(`#${redSide}PlayoffAllianceWins`).text(data.Matchup.RedAllianceWins);
      $(`#${blueSide}PlayoffAllianceWins`).text(data.Matchup.BlueAllianceWins);
      $("#playoffSeriesStatus").css("display", "flex");
    } else {
      $("#playoffSeriesStatus").hide();
    }
  } else {
    $(`#${redSide}PlayoffAlliance`).text("");
    $(`#${blueSide}PlayoffAlliance`).text("");
    $(".playoff-alliance").hide();
    $("#playoffSeriesStatus").hide();
  }

  let matchName = data.Match.LongName;
  if (data.Match.NameDetail !== "") {
    matchName += " &ndash; " + data.Match.NameDetail;
  }
  $("#matchName").html(matchName);
  $("#timeoutNextMatchName").html(matchName);
  $("#timeoutBreakDescription").text(data.BreakDescription);
};

// Handles a websocket message to update the match time countdown.
const handleMatchTime = function (data) {
  translateMatchTime(data, function (matchState, matchStateText, countdownSec) {
    $("#matchTime").text(getCountdownString(countdownSec));

    // Show/hide auto indicator during AUTO_PERIOD
    if (matchState === "AUTO_PERIOD" || matchState === "PAUSE_PERIOD") {
      $("#autoIndicator").css("display", "flex");
    } else {
      $("#autoIndicator").hide();
    }

    // Handle shift timer display during TELEOP_PERIOD
    if (matchState === "TELEOP_PERIOD" && data.CurrentShift > 0) {
      updateShiftTimer(data.CurrentShift, data.ShiftTimeRemaining);
      $("#shiftInfo").css("display", "flex");
    } else {
      $("#shiftInfo").hide();
    }

    // Hide match under review when match is active
    if (matchState === "AUTO_PERIOD" || matchState === "TELEOP_PERIOD") {
      $("#matchUnderReview").hide();
    }
  });
};

// Updates the shift timer display during teleop using server-provided shift data
const updateShiftTimer = function (currentShift, shiftTimeRemaining) {
  if (currentShift > 0) {
    $("#shiftProgress").text(`${currentShift} / 6`);
    const seconds = shiftTimeRemaining % 60;
    $("#shiftTimeRemaining").text(`:${seconds.toString().padStart(2, '0')}`);
  }
};

// Handles a websocket message to update the match score.
const handleRealtimeScore = function (data) {
  $(`#${redSide}ScoreNumber`).text(data.Red.ScoreSummary.Score);
  $(`#${blueSide}ScoreNumber`).text(data.Blue.ScoreSummary.Score);

  // Update fuel counters for ranking points
  updateFuelCounter(redSide, data.Red.Score.Hub.TotalFuel);
  updateFuelCounter(blueSide, data.Blue.Score.Hub.TotalFuel);

  // Update hub active indicators
  if (data.Red.Score.Hub.IsActive) {
    $(`#${redSide}HubActive`).css("visibility", "visible");
  } else {
    $(`#${redSide}HubActive`).css("visibility", "hidden");
  }

  if (data.Blue.Score.Hub.IsActive) {
    $(`#${blueSide}HubActive`).css("visibility", "visible");
  } else {
    $(`#${blueSide}HubActive`).css("visibility", "hidden");
  }
};

// Update fuel counter display based on current fuel and ranking point thresholds
const updateFuelCounter = function (side, totalFuel) {
  const firstThreshold = 100;
  const secondThreshold = 360;

  let currentTarget = firstThreshold;
  if (totalFuel >= firstThreshold) {
    currentTarget = secondThreshold;
  }

  $(`#${side}FuelCount`).text(totalFuel || 0);
  $(`#${side}FuelTarget`).text(currentTarget);
};

// Show match under review indicator
const handleMatchReview = function (isUnderReview) {
  if (isUnderReview) {
    $("#matchUnderReview").css("display", "flex");
    $("#matchTime").hide();
  } else {
    $("#matchUnderReview").hide();
    $("#matchTime").show();
  }
};

// Handles a websocket message to populate the final score data.
const handleScorePosted = function (data) {
  // Show event high score banner if this is a new high score
  if (data.IsEventHighScore) {
    $("#eventHighScore").css("display", "flex");
  } else {
    $("#eventHighScore").hide();
  }

  $(`#${redSide}FinalScore`).text(data.RedScoreSummary.Score);
  $(`#${redSide}FinalAlliance`).text("Alliance " + data.Match.PlayoffRedAlliance);
  setTeamInfo(redSide, 1, data.Match.Red1, data.RedCards, data.RedRankings);
  setTeamInfo(redSide, 2, data.Match.Red2, data.RedCards, data.RedRankings);
  setTeamInfo(redSide, 3, data.Match.Red3, data.RedCards, data.RedRankings);
  if (data.RedOffFieldTeamIds.length > 0) {
    setTeamInfo(redSide, 4, data.RedOffFieldTeamIds[0], data.RedCards, data.RedRankings);
  } else {
    setTeamInfo(redSide, 4, 0, data.RedCards, data.RedRankings);
  }
  $(`#${redSide}FinalAutoFuelPoints`).text(data.RedScoreSummary.AutoFuelPoints);
  $(`#${redSide}FinalAutoTowerPoints`).text(data.RedScoreSummary.AutoTowerPoints);
  $(`#${redSide}FinalTowerPoints`).text(data.RedScoreSummary.TowerPoints);
  $(`#${redSide}FinalMatchPoints`).text(data.RedScoreSummary.MatchPoints);
  $(`#${redSide}FinalFoulPoints`).text(data.RedScoreSummary.FoulPoints);

  // Use icons for ranking points instead of checkmarks
  // Trophy icons for win (3 RPs), ball icon for Energized RP, multi-ball for Supercharged RP, tower for Traversal RP
  $(`#${redSide}FinalEnergizedRankingPoint`).html(
    data.RedScoreSummary.EnergizedRankingPoint ? '<i class="bi bi-circle-fill"></i>' : '<i class="bi bi-circle"></i>'
  );
  $(`#${redSide}FinalEnergizedRankingPoint`).attr(
    "data-checked", data.RedScoreSummary.EnergizedRankingPoint
  );
  $(`#${redSide}FinalSuperchargedRankingPoint`).html(
    data.RedScoreSummary.SuperchargedRankingPoint ? '<i class="bi bi-circles"></i>' : '<i class="bi bi-circle"></i>'
  );
  $(`#${redSide}FinalSuperchargedRankingPoint`).attr(
    "data-checked", data.RedScoreSummary.SuperchargedRankingPoint
  );
  $(`#${redSide}FinalTraversalRankingPoint`).html(
    data.RedScoreSummary.TraversalRankingPoint ? '<i class="bi bi-building"></i>' : '<i class="bi bi-building"></i>'
  );
  $(`#${redSide}FinalTraversalRankingPoint`).attr(
    "data-checked", data.RedScoreSummary.TraversalRankingPoint
  );
  $(`#${redSide}FinalRankingPoints`).html(data.RedRankingPoints);
  $(`#${redSide}FinalWins`).text(data.RedWins);
  const redFinalDestination = $(`#${redSide}FinalDestination`);
  redFinalDestination.html(data.RedDestination.replace("Advances to ", "Advances to<br>"));
  redFinalDestination.toggle(data.RedDestination !== "");
  redFinalDestination.attr("data-won", data.RedWon);

  $(`#${blueSide}FinalScore`).text(data.BlueScoreSummary.Score);
  $(`#${blueSide}FinalAlliance`).text("Alliance " + data.Match.PlayoffBlueAlliance);
  setTeamInfo(blueSide, 1, data.Match.Blue1, data.BlueCards, data.BlueRankings);
  setTeamInfo(blueSide, 2, data.Match.Blue2, data.BlueCards, data.BlueRankings);
  setTeamInfo(blueSide, 3, data.Match.Blue3, data.BlueCards, data.BlueRankings);
  if (data.BlueOffFieldTeamIds.length > 0) {
    setTeamInfo(blueSide, 4, data.BlueOffFieldTeamIds[0], data.BlueCards, data.BlueRankings);
  } else {
    setTeamInfo(blueSide, 4, 0, data.BlueCards, data.BlueRankings);
  }
  $(`#${blueSide}FinalAutoFuelPoints`).text(data.BlueScoreSummary.AutoFuelPoints);
  $(`#${blueSide}FinalAutoTowerPoints`).text(data.BlueScoreSummary.AutoTowerPoints);
  $(`#${blueSide}FinalTowerPoints`).text(data.BlueScoreSummary.TowerPoints);
  $(`#${blueSide}FinalMatchPoints`).text(data.BlueScoreSummary.MatchPoints);
  $(`#${blueSide}FinalFoulPoints`).text(data.BlueScoreSummary.FoulPoints);
  $(`#${blueSide}FinalEnergizedRankingPoint`).html(
    data.BlueScoreSummary.EnergizedRankingPoint ? '<i class="bi bi-circle-fill"></i>' : '<i class="bi bi-circle"></i>'
  );
  $(`#${blueSide}FinalEnergizedRankingPoint`).attr(
    "data-checked", data.BlueScoreSummary.EnergizedRankingPoint
  );
  $(`#${blueSide}FinalSuperchargedRankingPoint`).html(
    data.BlueScoreSummary.SuperchargedRankingPoint ? '<i class="bi bi-circles"></i>' : '<i class="bi bi-circle"></i>'
  );
  $(`#${blueSide}FinalSuperchargedRankingPoint`).attr(
    "data-checked", data.BlueScoreSummary.SuperchargedRankingPoint
  );
  $(`#${blueSide}FinalTraversalRankingPoint`).html(
    data.BlueScoreSummary.TraversalRankingPoint ? '<i class="bi bi-building"></i>' : '<i class="bi bi-building"></i>'
  );
  $(`#${blueSide}FinalTraversalRankingPoint`).attr(
    "data-checked", data.BlueScoreSummary.TraversalRankingPoint
  );
  $(`#${blueSide}FinalRankingPoints`).html(data.BlueRankingPoints);
  $(`#${blueSide}FinalWins`).text(data.BlueWins);
  const blueFinalDestination = $(`#${blueSide}FinalDestination`);
  blueFinalDestination.html(data.BlueDestination.replace("Advances to ", "Advances to<br>"));
  blueFinalDestination.toggle(data.BlueDestination !== "");
  blueFinalDestination.attr("data-won", data.BlueWon);

  let matchName = data.Match.LongName;
  if (data.Match.NameDetail !== "") {
    matchName += " &ndash; " + data.Match.NameDetail;
  }
  $("#finalMatchName").html(matchName);

  // Reload the bracket to reflect any changes.
  $("#bracketSvg").attr("src", "/api/bracket/svg?activeMatch=saved&v=" + new Date().getTime());

  if (data.Match.Type === matchTypePlayoff) {
    // Hide bonus ranking points and show playoff-only fields.
    $(".playoff-hidden-field").hide();
    $(".playoff-only-field").show();
  } else {
    $(".playoff-hidden-field").show();
    $(".playoff-only-field").hide();
  }
  $(".coopertition-hidden-field").toggle(data.CoopertitionEnabled);
};

// Handles a websocket message to play a sound to signal match start/stop/etc.
const handlePlaySound = function (sound) {
  $("audio").each(function (k, v) {
    // Stop and reset any sounds that are still playing.
    v.pause();
    v.currentTime = 0;
  });
  $("#sound-" + sound)[0].play();
};

// Handles a websocket message to update the alliance selection screen.
const handleAllianceSelection = function (data) {
  const alliances = data.Alliances;
  const rankedTeams = data.RankedTeams;
  if (alliances && alliances.length > 0) {
    const numColumns = alliances[0].TeamIds.length + 1;
    $.each(alliances, function (k, v) {
      v.Index = k + 1;
    });
    $("#allianceSelection").html(allianceSelectionTemplate({alliances: alliances, numColumns: numColumns}));

    // Apply strikethrough for declined teams and highlighting for captains
    if (data.DeclinedTeamIds) {
      $.each(data.DeclinedTeamIds, function (i, teamId) {
        $(`.selection-cell:contains(${teamId})`).attr("data-declined", "true");
      });
    }

    // Highlight alliance captains (first team in each alliance)
    $(".selection-cell").each(function(index) {
      // Every 4th cell starting from index 1 is a captain (alliance structure: alliance#, captain, pick1, pick2, pick3)
      if (index % 4 === 1 && $(this).text().trim() !== "") {
        $(this).attr("data-captain", "true");
      }
    });
  }
  if (rankedTeams) {
    let text = "";
    $.each(rankedTeams, function (i, v) {
      if (!v.Picked) {
        text += `<div class="unpicked"><div class="unpicked-rank">${v.Rank}.</div>` +
          `<div class="unpicked-team">${v.TeamId}</div></div>`;
      }
    });
    $("#allianceRankings").html(text);
  }

  if (data.ShowTimer) {
    $("#allianceSelectionTimer").text(getCountdownString(data.TimeRemainingSec));
  } else {
    $("#allianceSelectionTimer").html("&nbsp;");
  }
};

// Handles a websocket message to populate and/or show/hide a lower third.
const handleLowerThird = function (data) {
  if (data.LowerThird !== null) {
    if (data.LowerThird.BottomText === "") {
      $("#lowerThirdTop").hide();
      $("#lowerThirdBottom").hide();
      $("#lowerThirdSingle").text(data.LowerThird.TopText);
      $("#lowerThirdSingle").show();
    } else {
      $("#lowerThirdSingle").hide();
      $("#lowerThirdTop").text(data.LowerThird.TopText);
      $("#lowerThirdBottom").text(data.LowerThird.BottomText);
      $("#lowerThirdTop").show();
      $("#lowerThirdBottom").show();
    }
  }

  const lowerThirdElement = $("#lowerThird");
  if (data.ShowLowerThird && !lowerThirdElement.is(":visible")) {
    lowerThirdElement.show();
    lowerThirdElement.transition({queue: false, left: "150px"}, 750, "ease");
  } else if (!data.ShowLowerThird && lowerThirdElement.is(":visible")) {
    lowerThirdElement.transition({queue: false, left: "-1000px"}, 1000, "ease", function () {
      lowerThirdElement.hide();
    });
  }
};

const transitionAllianceSelectionToBlank = function (callback) {
  $('#allianceSelectionCentering').transition({queue: false, right: "-60em"}, 500, "ease", callback);
  $('#allianceRankingsCentering.enabled').transition({queue: false, left: "-60em"}, 500, "ease");
};

const transitionBlankToAllianceSelection = function (callback) {
  $('#allianceSelectionCentering').css("right", "-60em").show();
  $('#allianceSelectionCentering').transition({queue: false, right: "3em"}, 500, "ease", callback);
  $('#allianceRankingsCentering.enabled').css("left", "-60em").show();
  $('#allianceRankingsCentering.enabled').transition({queue: false, left: "3em"}, 500, "ease");
};

const transitionBlankToBracket = function (callback) {
  transitionBlankToLogo(function () {
    setTimeout(function () {
      transitionLogoToBracket(callback);
    }, 50);
  });
};

const transitionBlankToIntro = function (callback) {
  // Show the Rebuilt scorebar with animation (intro shows only header, team numbers, and avatars)
  // Fade out scores, timer, and stats for intro
  $("#centerClock").transition({queue: false, opacity: 0}, 300, "ease", function () {
    $("#centerClock").hide();
  });
  $("#leftScoreNumber").transition({queue: false, opacity: 0}, 300, "ease", function () {
    $("#leftScoreNumber").hide();
  });
  $("#rightScoreNumber").transition({queue: false, opacity: 0}, 300, "ease", function () {
    $("#rightScoreNumber").hide();
  });
  $(".stats-block").transition({queue: false, opacity: 0}, 300, "ease", function () {
    $(".stats-block").hide();
  });

  $("#matchOverlay").css({opacity: 0, display: "block"});
  $("#audienceHeader").css({opacity: 0, transform: "translateY(-100%)", display: "flex"});
  $("#audienceScorebar").css({opacity: 0, transform: "translateY(100%)", display: "flex"});

  $("#matchOverlay").transition({queue: false, opacity: 1}, 500, "ease");
  $("#audienceHeader").transition({queue: false, opacity: 1, translateY: 0}, 500, "ease");
  $("#audienceScorebar").transition({queue: false, opacity: 1, translateY: 0}, 500, "ease", callback);
};

const transitionBlankToLogo = function (callback) {
  $(".blindsCenter.blank").css({rotateY: "0deg"});
  $(".blindsCenter.full").css({rotateY: "-180deg"});
  $(".blinds.right").transition({queue: false, right: 0}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: 0}, 1000, "ease", function () {
    $(".blinds.left").addClass("full");
    $(".blinds.right").hide();
    setTimeout(function () {
      $(".blindsCenter.blank").transition({queue: false, rotateY: "180deg"}, 500, "ease");
      $(".blindsCenter.full").transition({queue: false, rotateY: "0deg"}, 500, "ease", callback);
    }, 200);
  });
};

const transitionBlankToLogoLuma = function (callback) {
  $(".blindsCenter.blank").css({rotateY: "180deg"});
  $(".blindsCenter.full").transition({queue: false, rotateY: "0deg"}, 1000, "ease", callback);
};

const transitionBlankToMatch = function (callback) {
  // Show the Rebuilt scorebar with animation
  // Make sure all match play elements are visible
  $("#centerClock").show();
  $("#leftScoreNumber").show();
  $("#rightScoreNumber").show();
  $(".stats-block").show();

  $("#matchOverlay").css({opacity: 0, display: "block"});
  $("#audienceHeader").css({opacity: 0, transform: "translateY(-100%)", display: "flex"});
  $("#audienceScorebar").css({opacity: 0, transform: "translateY(100%)", display: "flex"});

  $("#matchOverlay").transition({queue: false, opacity: 1}, 500, "ease");
  $("#audienceHeader").transition({queue: false, opacity: 1, translateY: 0}, 500, "ease");
  $("#audienceScorebar").transition({queue: false, opacity: 1, translateY: 0}, 500, "ease", callback);
};

const transitionBlankToScore = function (callback) {
  transitionBlankToLogo(function () {
    setTimeout(function () {
      transitionLogoToScore(callback);
    }, 50);
  });
};

const transitionBlankToSponsor = function (callback) {
  $(".blindsCenter.blank").css({rotateY: "90deg"});
  $(".blinds.right").transition({queue: false, right: 0}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: 0}, 1000, "ease", function () {
    $(".blinds.left").addClass("full");
    $(".blinds.right").hide();
    setTimeout(function () {
      $("#sponsor").show();
      $("#sponsor").transition({queue: false, opacity: 1}, 1000, "ease", callback);
    }, 200);
  });
};

const transitionBlankToTimeout = function (callback) {
  $("#overlayCentering").transition(overlayCenteringShowParams, 500, "ease", function () {
    $("#timeoutDetails").transition({queue: false, width: timeoutDetailsOut}, 500, "ease");
    $("#logo").transition({queue: false, top: logoUp}, 500, "ease", function () {
      $(".timeout-detail").transition({queue: false, opacity: 1}, 750, "ease");
      $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
    });
  });
};

const transitionBracketToBlank = function (callback) {
  transitionBracketToLogo(function () {
    transitionLogoToBlank(callback);
  });
};

const transitionBracketToLogo = function (callback) {
  $("#bracket").transition({queue: false, opacity: 0}, 500, "ease", function () {
    $("#bracket").hide();
  });
  $(".blindsCenter.full").transition({queue: false, top: 0, scale: 1}, 625, "ease", callback);
};

const transitionBracketToLogoLuma = function (callback) {
  transitionBracketToLogo(function () {
    transitionLogoToLogoLuma(callback);
  });
};

const transitionBracketToScore = function (callback) {
  $(".blindsCenter.full").transition({queue: false, top: scoreLogoTop, scale: 1}, 1000, "ease");
  $("#bracket").transition({queue: false, opacity: 0}, 1000, "ease", function () {
    $("#bracket").hide();
    $("#finalScore").show();
    $("#finalScore").transition({queue: false, opacity: 1}, 1000, "ease", callback);
  });
};

const transitionBracketToSponsor = function (callback) {
  transitionBracketToLogo(function () {
    transitionLogoToSponsor(callback);
  });
};

const transitionIntroToBlank = function (callback) {
  // Hide the Rebuilt scorebar with animation
  $("#audienceHeader").transition({queue: false, opacity: 0, translateY: "-100%"}, 500, "ease");
  $("#audienceScorebar").transition({queue: false, opacity: 0, translateY: "100%"}, 500, "ease");
  $("#matchOverlay").transition({queue: false, opacity: 0}, 500, "ease", function () {
    $("#audienceHeader").css("display", "none");
    $("#audienceScorebar").css("display", "none");
    $("#matchOverlay").css("display", "none");
    callback();
  });
};

const transitionIntroToMatch = function (callback) {
  // Fade in the elements that were hidden for intro: scores, timer, and stats
  $("#centerClock").show();
  $("#leftScoreNumber").show();
  $("#rightScoreNumber").show();
  $(".stats-block").show();

  $("#centerClock").transition({queue: false, opacity: 1}, 300, "ease");
  $("#leftScoreNumber").transition({queue: false, opacity: 1}, 300, "ease");
  $("#rightScoreNumber").transition({queue: false, opacity: 1}, 300, "ease");
  $(".stats-block").transition({queue: false, opacity: 1}, 300, "ease", callback);
};

const transitionMatchToIntro = function (callback) {
  // Fade out the match play elements for intro display
  $("#centerClock").transition({queue: false, opacity: 0}, 300, "ease", function () {
    $("#centerClock").hide();
  });
  $("#leftScoreNumber").transition({queue: false, opacity: 0}, 300, "ease", function () {
    $("#leftScoreNumber").hide();
  });
  $("#rightScoreNumber").transition({queue: false, opacity: 0}, 300, "ease", function () {
    $("#rightScoreNumber").hide();
  });
  $(".stats-block").transition({queue: false, opacity: 0}, 300, "ease", function () {
    $(".stats-block").hide();
    callback();
  });
};

const transitionIntroToTimeout = function (callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease", function () {
    $("#eventMatchInfo").hide();
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function () {
      $(".avatars").css("opacity", 0);
      $(".avatars").hide();
      $(".teams").hide();
      $("#timeoutDetails").transition({queue: false, width: timeoutDetailsOut}, 500, "ease");
      $("#logo").transition({queue: false, top: logoUp}, 500, "ease", function () {
        $(".timeout-detail").transition({queue: false, opacity: 1}, 750, "ease");
        $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
      });
    });
  });
};

const transitionLogoToBlank = function (callback) {
  $(".blindsCenter.blank").transition({queue: false, rotateY: "360deg"}, 500, "ease");
  $(".blindsCenter.full").transition({queue: false, rotateY: "180deg"}, 500, "ease", function () {
    setTimeout(function () {
      $(".blinds.left").removeClass("full");
      $(".blinds.right").show();
      $(".blinds.right").transition({queue: false, right: "-50%"}, 1000, "ease");
      $(".blinds.left").transition({queue: false, left: "-50%"}, 1000, "ease", callback);
    }, 200);
  });
};

const transitionLogoToBracket = function (callback) {
  $(".blindsCenter.full").transition({queue: false, top: bracketLogoTop, scale: bracketLogoScale}, 625, "ease");
  $("#bracket").show();
  $("#bracket").transition({queue: false, opacity: 1}, 1000, "ease", callback);
};

const transitionLogoToLogoLuma = function (callback) {
  $(".blinds.left").removeClass("full");
  $(".blinds.right").show();
  $(".blinds.right").transition({queue: false, right: "-50%"}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: "-50%"}, 1000, "ease", function () {
    if (callback) {
      callback();
    }
  });
};

const transitionLogoToScore = function (callback) {
  $(".blindsCenter.full").transition({queue: false, top: scoreLogoTop}, 625, "ease");
  $("#finalScore").show();
  $("#finalScore").transition({queue: false, opacity: 1}, 1000, "ease", callback);
};

const transitionLogoToSponsor = function (callback) {
  $(".blindsCenter.full").transition({queue: false, rotateY: "90deg"}, 750, "ease", function () {
    $("#sponsor").show();
    $("#sponsor").transition({queue: false, opacity: 1}, 1000, "ease", callback);
  });
};

const transitionLogoLumaToBlank = function (callback) {
  $(".blindsCenter.full").transition({queue: false, rotateY: "180deg"}, 1000, "ease", callback);
};

const transitionLogoLumaToBracket = function (callback) {
  transitionLogoLumaToLogo(function () {
    transitionLogoToBracket(callback);
  });
};

const transitionLogoLumaToLogo = function (callback) {
  $(".blinds.right").transition({queue: false, right: 0}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: 0}, 1000, "ease", function () {
    $(".blinds.left").addClass("full");
    $(".blinds.right").hide();
    if (callback) {
      callback();
    }
  });
};

const transitionLogoLumaToScore = function (callback) {
  transitionLogoLumaToLogo(function () {
    transitionLogoToScore(callback);
  });
};

const transitionMatchToBlank = function (callback) {
  // Hide the Rebuilt scorebar with animation
  $("#audienceHeader").transition({queue: false, opacity: 0, translateY: "-100%"}, 500, "ease");
  $("#audienceScorebar").transition({queue: false, opacity: 0, translateY: "100%"}, 500, "ease");
  $("#matchOverlay").transition({queue: false, opacity: 0}, 500, "ease", function () {
    $("#audienceHeader").css({display: "none", opacity: "", transform: ""});
    $("#audienceScorebar").css({display: "none", opacity: "", transform: ""});
    $("#matchOverlay").css({display: "none", opacity: ""});
    callback();
  });
};

const transitionScoreToBlank = function (callback) {
  transitionScoreToLogo(function () {
    transitionLogoToBlank(callback);
  });
};

const transitionScoreToBracket = function (callback) {
  $(".blindsCenter.full").transition({queue: false, top: bracketLogoTop, scale: bracketLogoScale}, 1000, "ease");
  $("#finalScore").transition({queue: false, opacity: 0}, 1000, "ease", function () {
    $("#finalScore").hide();
    $("#bracket").show();
    $("#bracket").transition({queue: false, opacity: 1}, 1000, "ease", callback);
  });
};

const transitionScoreToLogo = function (callback) {
  $("#finalScore").transition({queue: false, opacity: 0}, 500, "ease", function () {
    $("#finalScore").hide();
  });
  $(".blindsCenter.full").transition({queue: false, top: 0}, 625, "ease", callback);
};

const transitionScoreToLogoLuma = function (callback) {
  transitionScoreToLogo(function () {
    transitionLogoToLogoLuma(callback);
  });
};

const transitionScoreToSponsor = function (callback) {
  transitionScoreToLogo(function () {
    transitionLogoToSponsor(callback);
  });
};

const transitionSponsorToBlank = function (callback) {
  $("#sponsor").transition({queue: false, opacity: 0}, 1000, "ease", function () {
    setTimeout(function () {
      $(".blinds.left").removeClass("full");
      $(".blinds.right").show();
      $(".blinds.right").transition({queue: false, right: "-50%"}, 1000, "ease");
      $(".blinds.left").transition({queue: false, left: "-50%"}, 1000, "ease", callback);
      $("#sponsor").hide();
    }, 200);
  });
};

const transitionSponsorToBracket = function (callback) {
  transitionSponsorToLogo(function () {
    transitionLogoToBracket(callback);
  });
};

const transitionSponsorToLogo = function (callback) {
  $("#sponsor").transition({queue: false, opacity: 0}, 1000, "ease", function () {
    $(".blindsCenter.full").transition({queue: false, rotateY: "0deg"}, 750, "ease", callback);
    $("#sponsor").hide();
  });
};

const transitionSponsorToScore = function (callback) {
  transitionSponsorToLogo(function () {
    transitionLogoToScore(callback);
  });
};

const transitionTimeoutToBlank = function (callback) {
  $(".timeout-detail").transition({queue: false, opacity: 0}, 300, "linear");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function () {
    $("#timeoutDetails").transition({queue: false, width: timeoutDetailsIn}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease", function () {
      $("#overlayCentering").transition(overlayCenteringHideParams, 1000, "ease", callback);
    });
  });
};

const transitionTimeoutToIntro = function (callback) {
  $(".timeout-detail").transition({queue: false, opacity: 0}, 300, "linear");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function () {
    $("#timeoutDetails").transition({queue: false, width: timeoutDetailsIn}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease", function () {
      $(".avatars").css("display", "flex");
      $(".avatars").css("opacity", 1);
      $(".teams").css("display", "flex");
      $(".score").transition({queue: false, width: scoreMid}, 500, "ease", function () {
        $("#eventMatchInfo").show();
        $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoDown}, 500, "ease", callback);
      });
    });
  });
};

// Loads sponsor slide data and builds the slideshow HTML.
const initializeSponsorDisplay = function () {
  $.getJSON("/api/sponsor_slides", function (slides) {
    $("#sponsorContainer").empty();

    // Inject the HTML for each slide into the DOM.
    $.each(slides, function (index, slide) {
      slide.DisplayTimeMs = slide.DisplayTimeSec * 1000;
      slide.First = index === 0;

      let slideHtml;
      if (slide.Image) {
        slideHtml = sponsorImageTemplate(slide);
      } else {
        slideHtml = sponsorTextTemplate(slide);
      }
      $("#sponsorContainer").append(slideHtml);
    });
  });
};

const getAvatarUrl = function (teamId) {
  return "/api/teams/" + teamId + "/avatar";
};

const setTeamInfo = function (side, position, teamId, cards, rankings) {
  const teamNumberElement = $(`#${side}FinalTeam${position}`);
  teamNumberElement.html(teamId);
  teamNumberElement.toggle(teamId > 0);
  const avatarElement = $(`#${side}FinalTeam${position}Avatar`);
  avatarElement.attr("src", getAvatarUrl(teamId));
  avatarElement.toggle(teamId > 0);

  const cardElement = $(`#${side}FinalTeam${position}Card`);
  cardElement.attr("data-card", cards[teamId.toString()] || "");

  const ranking = rankings[teamId];
  let rankIndicator = "";
  let rankNumber = "";
  if (ranking !== undefined && ranking !== null && ranking.Rank !== 0) {
    rankNumber = ranking.Rank;
    if (rankNumber > ranking.PreviousRank && ranking.PreviousRank > 0) {
      rankIndicator = "rank-down";
    } else if (rankNumber < ranking.PreviousRank) {
      rankIndicator = "rank-up";
    }
  }

  const rankIndicatorElement = $(`#${side}FinalTeam${position}RankIndicator`);
  rankIndicatorElement.attr("src", rankIndicator === "" ? "" : `/static/img/${rankIndicator}.svg`);
  rankIndicatorElement.toggle(rankIndicator !== "" && teamId > 0);

  const rankNumberElement = $(`#${side}FinalTeam${position}RankNumber`);
  rankNumberElement.text(rankNumber);
  rankNumberElement.toggle(teamId > 0);
};

$(function () {
  // Read the configuration for this display from the URL query string.
  const urlParams = new URLSearchParams(window.location.search);
  document.body.style.backgroundColor = urlParams.get("background");
  const reversed = urlParams.get("reversed");
  if (reversed === "true") {
    redSide = "right";
    blueSide = "left";
  } else {
    redSide = "left";
    blueSide = "right";
  }
  $(".reversible-left").attr("data-reversed", reversed);
  $(".reversible-right").attr("data-reversed", reversed);
  if (urlParams.get("overlayLocation") === "top") {
    overlayCenteringHideParams = overlayCenteringTopHideParams;
    overlayCenteringShowParams = overlayCenteringTopShowParams;
    $("#overlayCentering").css("top", overlayCenteringTopUp);
  } else {
    overlayCenteringHideParams = overlayCenteringBottomHideParams;
    overlayCenteringShowParams = overlayCenteringBottomShowParams;
  }

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/audience/websocket", {
    allianceSelection: function (event) {
      handleAllianceSelection(event.data);
    },
    audienceDisplayMode: function (event) {
      handleAudienceDisplayMode(event.data);
    },
    lowerThird: function (event) {
      handleLowerThird(event.data);
    },
    matchLoad: function (event) {
      handleMatchLoad(event.data);
    },
    matchTime: function (event) {
      handleMatchTime(event.data);
    },
    matchTiming: function (event) {
      handleMatchTiming(event.data);
    },
    playSound: function (event) {
      handlePlaySound(event.data);
    },
    realtimeScore: function (event) {
      handleRealtimeScore(event.data);
    },
    scorePosted: function (event) {
      handleScorePosted(event.data);
    },
  });

  // Map how to transition from one screen to another. Missing links between screens indicate that first we
  // must transition to the blank screen and then to the target screen.
  transitionMap = {
    allianceSelection: {
      blank: transitionAllianceSelectionToBlank,
    },
    blank: {
      allianceSelection: transitionBlankToAllianceSelection,
      bracket: transitionBlankToBracket,
      intro: transitionBlankToIntro,
      logo: transitionBlankToLogo,
      logoLuma: transitionBlankToLogoLuma,
      match: transitionBlankToMatch,
      score: transitionBlankToScore,
      sponsor: transitionBlankToSponsor,
      timeout: transitionBlankToTimeout,
    },
    bracket: {
      blank: transitionBracketToBlank,
      logo: transitionBracketToLogo,
      logoLuma: transitionBracketToLogoLuma,
      score: transitionBracketToScore,
      sponsor: transitionBracketToSponsor,
    },
    intro: {
      blank: transitionIntroToBlank,
      match: transitionIntroToMatch,
      timeout: transitionIntroToTimeout,
    },
    logo: {
      blank: transitionLogoToBlank,
      bracket: transitionLogoToBracket,
      logoLuma: transitionLogoToLogoLuma,
      score: transitionLogoToScore,
      sponsor: transitionLogoToSponsor,
    },
    logoLuma: {
      blank: transitionLogoLumaToBlank,
      bracket: transitionLogoLumaToBracket,
      logo: transitionLogoLumaToLogo,
      score: transitionLogoLumaToScore,
    },
    match: {
      blank: transitionMatchToBlank,
      intro: transitionMatchToIntro,
    },
    score: {
      blank: transitionScoreToBlank,
      bracket: transitionScoreToBracket,
      logo: transitionScoreToLogo,
      logoLuma: transitionScoreToLogoLuma,
      sponsor: transitionScoreToSponsor,
    },
    sponsor: {
      blank: transitionSponsorToBlank,
      bracket: transitionSponsorToBracket,
      logo: transitionSponsorToLogo,
      score: transitionSponsorToScore,
    },
    timeout: {
      blank: transitionTimeoutToBlank,
      intro: transitionTimeoutToIntro,
    },
  }
});
