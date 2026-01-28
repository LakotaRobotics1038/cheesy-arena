// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model of a game-specific rule.

package game

type Rule struct {
	Id             int
	RuleNumber     string
	IsMajor        bool
	IsRankingPoint bool
	Description    string
}

// All rules from the 2026 REBUILT game that carry point penalties.
// @formatter:off
var rules = []*Rule{
	{1, "G206", false, true, "A team or ALLIANCE may not collude with another team to each purposefully violate a rule in an attempt to influence Ranking Points."},
	{2, "G210", true, false, "A strategy not consistent with standard gameplay and clearly aimed at forcing the opponent ALLIANCE to violate a rule is not in the spirit of FIRST Robotics Competition and not allowed."},
	{3, "G301", true, false, "A DRIVE TEAM member may not cause significant delays to the start of their MATCH."},
	{4, "G401", false, false, "In AUTO, each DRIVE TEAM member must remain in their staged areas. A DRIVE TEAM member staged behind a HUMAN STARTING LINE may not contact anything in front of that HUMAN STARTING LINE, unless for personal or equipment safety, to press the E-Stop or A-Stop, or granted permission by a Head REFEREE or FTA."},
	{5, "G402", false, true, "In AUTO, a DRIVE TEAM member may not directly or indirectly interact with a ROBOT or an OPERATOR CONSOLE unless for personal safety, OPERATOR CONSOLE safety, or pressing an E-Stop or A-Stop. A HUMAN PLAYER entering FUEL onto the FIELD is an exception to this rule."},
	{6, "G403", true, false, "In AUTO, a ROBOT whose BUMPERS are completely across the CENTER LINE may not contact an opponent ROBOT."},
	{7, "G404", true, false, "A ROBOT may not deliberately use a SCORING ELEMENT in an attempt to ease or amplify a challenge associated with a FIELD element."},
	{8, "G405", false, false, "A ROBOT may not intentionally eject SCORING ELEMENTS from the FIELD with an exception of the OUTPOST."},
	{9, "G406", false, false, "Neither a ROBOT nor a HUMAN PLAYER may damage a SCORING ELEMENT."},
	{10, "G407", true, false, "A ROBOT may not launch a SCORING ELEMENT into their HUB unless their BUMPERS are partially or fully within their ALLIANCE ZONE."},
	{11, "G408", false, false, "A ROBOT may not do either of the following with FUEL released by the HUB unless and until that FUEL contacts anything else: A. gain greater than MOMENTARY CONTROL of FUEL, or B. push or redirect FUEL to a desired location."},
	{12, "G409", true, false, "A ROBOT may not pose an undue hazard to a human, an ARENA element, or another ROBOT."},
	{13, "G410", false, false, "ROBOT extensions may not interact with the carpet, BUMPS, or TOWER BASE such that the BUMPERS are lifted out of the BUMPER ZONE."},
	{14, "G411", false, false, "A ROBOT may not damage FIELD elements."},
	{15, "G412", true, false, "A ROBOT is prohibited from the following interactions with FIELD elements (with exception of RUNGS and UPRIGHTS): A. grabbing, B. grasping, C. attaching to, D. becoming entangled with, E. suspending from."},
	{16, "G413", false, false, "A ROBOT may not extend beyond any of the horizontal or vertical expansion limits."},
	{17, "G414", false, false, "ROBOTS may not fully support the weight of other ROBOTS on their ALLIANCE to climb the TOWER."},
	{18, "G415", false, false, "A ROBOT may not use a COMPONENT outside its ROBOT PERIMETER (except its BUMPERS) to initiate contact with an opponent ROBOT inside the vertical projection of the opponent's ROBOT PERIMETER."},
	{19, "G416", true, false, "A ROBOT may not damage or functionally impair an opponent ROBOT deliberately or by initiating contact inside the vertical projection of an opponent's ROBOT PERIMETER."},
	{20, "G417", true, true, "A ROBOT may not deliberately attach to, tip, or entangle with an opponent ROBOT."},
	{21, "G418", false, false, "A ROBOT may not PIN an opponent's ROBOT for more than 3 seconds."},
	{22, "G419", true, false, "2 or more ROBOTS that appear to a REFEREE to be working together may not isolate or close off any major element of MATCH play."},
	{23, "G420", true, false, "A ROBOT may not contact, directly or transitively through a SCORING ELEMENT, an opponent ROBOT in contact with an opponent TOWER during the last 30 seconds of the MATCH."},
	{24, "G421", false, false, "A DRIVE TEAM member must remain in their designated area as follows: A. DRIVERS and DRIVE COACHES may not contact anything outside their ALLIANCE AREA, B. a DRIVER must use the OPERATOR CONSOLE in the DRIVER STATION to which they are assigned, C. a HUMAN PLAYER may not contact anything outside their ALLIANCE AREA, D. a TECHNICIAN may not contact anything outside their designated area."},
	{25, "G422", true, false, "A ROBOT shall be operated only by the DRIVERS and/or HUMAN PLAYERS of that team. A DRIVE COACH activating their E-Stop or A-Stop is the exception to this rule."},
	{26, "G423", false, false, "A DRIVE TEAM member may not extend into the CHUTE beyond the ALLIANCE-colored tape line while the CHUTE DOOR is open, or into the CORRAL beyond the ALLIANCE-colored tape line."},
	{27, "G424", true, false, "A DRIVE TEAM member may not deliberately use a SCORING ELEMENT in an attempt to ease or amplify a challenge associated with a FIELD element."},
	{28, "G425", true, false, "FUEL may only be introduced to the FIELD by a HUMAN PLAYER or DRIVER through the CHUTE, through the bottom opening in the OUTPOST, or thrown from the OUTPOST AREA."},
	{29, "G426", false, false, "DRIVE COACHES may not touch SCORING ELEMENTS, unless for safety purposes."},
	{30, "G427", false, false, "Off-FIELD FUEL may only be stored in the CHUTE and the CORRAL. Excess FUEL must immediately be entered onto the FIELD."},
}

// @formatter:on
var ruleMap map[int]*Rule

// Returns the rule having the given ID, or nil if no such rule exists.
func GetRuleById(id int) *Rule {
	return GetAllRules()[id]
}

// Returns a slice of all defined rules that carry point penalties.
func GetAllRules() map[int]*Rule {
	if ruleMap == nil {
		ruleMap = make(map[int]*Rule, len(rules))
		for _, rule := range rules {
			ruleMap[rule.Id] = rule
		}
	}
	return ruleMap
}
