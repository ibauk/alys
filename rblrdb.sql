BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "rallyparams" (
	"RallyTitle"	TEXT,
	"RallySlogan"	TEXT,
	"MaxHours"	INTEGER NOT NULL DEFAULT 0,
	"StartTime"	TEXT,
	"FinishTime"	TEXT,
	"MinMiles"	INTEGER NOT NULL DEFAULT 0,
	"PenaltyMaxMiles"	INTEGER NOT NULL DEFAULT 0,
	"MaxMilesMethod"	INTEGER NOT NULL DEFAULT 0,
	"MaxMilesPoints"	INTEGER NOT NULL DEFAULT 0,
	"PenaltyMilesDNF"	INTEGER NOT NULL DEFAULT 0,
	"MinPoints"	INTEGER NOT NULL DEFAULT 0,
	"ScoringMethod"	INTEGER NOT NULL DEFAULT 3,
	"ShowMultipliers"	INTEGER NOT NULL DEFAULT 2,
	"TiedPointsRanking"	INTEGER NOT NULL DEFAULT 1,
	"TeamRanking"	INTEGER NOT NULL DEFAULT 3,
	"OdoCheckMiles"	NUMERIC DEFAULT 0,
	"Cat1Label"	TEXT,
	"Cat2Label"	TEXT,
	"Cat3Label"	TEXT,
	"Cat4Label"	TEXT,
	"Cat5Label"	TEXT,
	"Cat6Label"	TEXT,
	"Cat7Label"	TEXT,
	"Cat8Label"	TEXT,
	"Cat9Label"	TEXT,
	"RejectReasons"	TEXT,
	"DBState"	INTEGER NOT NULL DEFAULT 0,
	"DBVersion"	INTEGER NOT NULL DEFAULT 12,
	"AutoRank"	INTEGER NOT NULL DEFAULT 1,
	"Theme"	TEXT NOT NULL DEFAULT 'default',
	"MilesKms"	INTEGER NOT NULL DEFAULT 0,
	"LocalTZ"	TEXT NOT NULL DEFAULT 'Europe/London',
	"DecimalComma"	INTEGER NOT NULL DEFAULT 0,
	"HostCountry"	TEXT NOT NULL DEFAULT 'UK',
	"Locale"	TEXT NOT NULL DEFAULT 'en-GB',
	"EmailParams"	TEXT,
	"isvirtual"	INTEGER NOT NULL DEFAULT 0,
	"tankrange"	INTEGER NOT NULL DEFAULT 200,
	"refuelstops"	TEXT,
	"stopmins"	INTEGER NOT NULL DEFAULT 10,
	"spbonus"	TEXT,
	"fpbonus"	TEXT,
	"mpbonus"	TEXT,
	"settings"	TEXT,
	"StartOption"	INTEGER NOT NULL DEFAULT 0,
	"ebcsettings"	TEXT,
	"CurrentLeg"	INTEGER NOT NULL DEFAULT 1,
	"NumLegs"	INTEGER NOT NULL DEFAULT 1,
	"LegData"	TEXT
);
CREATE TABLE IF NOT EXISTS "cohorts" (
	"Cohort"	INTEGER NOT NULL,
	"FixedStart"	INTEGER NOT NULL DEFAULT 1,
	"StartTime"	TEXT,
	PRIMARY KEY("Cohort")
);
CREATE TABLE IF NOT EXISTS "teams" (
	"TeamID"	INTEGER NOT NULL,
	"BriefDesc"	TEXT,
	PRIMARY KEY("TeamID")
);
CREATE TABLE IF NOT EXISTS "functions" (
	"functionid"	INTEGER,
	"menulbl"	TEXT,
	"url"	TEXT,
	"onclick"	TEXT,
	"Tags"	TEXT,
	PRIMARY KEY("functionid")
);
CREATE TABLE IF NOT EXISTS "menus" (
	"menuid"	TEXT,
	"menulbl"	TEXT,
	"menufuncs"	TEXT,
	PRIMARY KEY("menuid")
);
CREATE TABLE IF NOT EXISTS "certificates" (
	"EntrantID"	INTEGER NOT NULL DEFAULT 0,
	"css"	TEXT,
	"html"	TEXT,
	"options"	TEXT,
	"image"	TEXT,
	"Class"	INTEGER NOT NULL DEFAULT 0,
	"Title"	TEXT,
	PRIMARY KEY("EntrantID","Class")
);
CREATE TABLE IF NOT EXISTS "importspecs" (
	"specid"	TEXT NOT NULL,
	"specTitle"	TEXT,
	"importType"	INTEGER NOT NULL DEFAULT 0,
	"fieldSpecs"	TEXT,
	PRIMARY KEY("specid")
);
CREATE TABLE IF NOT EXISTS "themes" (
	"Theme"	TEXT NOT NULL,
	"css"	TEXT NOT NULL,
	PRIMARY KEY("Theme")
);
CREATE TABLE IF NOT EXISTS "classes" (
	"Class"	INTEGER NOT NULL,
	"BriefDesc"	TEXT NOT NULL,
	"AutoAssign"	INTEGER NOT NULL DEFAULT 1,
	"MinPoints"	INTEGER NOT NULL DEFAULT 0,
	"MinBonuses"	INTEGER NOT NULL DEFAULT 0,
	"BonusesReqd"	TEXT,
	"LowestRank"	INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY("Class")
);
CREATE TABLE IF NOT EXISTS "config" (
	"DBInitialised"	INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS "entrants" (
	"EntrantID"	INTEGER,
	"Bike"	TEXT,
	"BikeReg"	TEXT,
	"RiderName"	TEXT,
	"RiderFirst"	TEXT,
	"RiderLast"	TEXT,
	"RiderIBA"	INTEGER,
	"PillionName"	TEXT,
	"PillionFirst"	TEXT,
	"PillionLast"	TEXT,
	"PillionIBA"	INTEGER,
	"TeamID"	INTEGER NOT NULL DEFAULT 0,
	"Country"	TEXT DEFAULT 'UK',
	"OdoKms"	INTEGER NOT NULL DEFAULT 0,
	"OdoStart"	INTEGER,
	"OdoFinish"	INTEGER,
	"CorrectedMiles"	NUMERIC DEFAULT 0,
	"FinishTime"	TEXT,
	"StartTime"	TEXT,
	"EntrantStatus"	INTEGER NOT NULL DEFAULT 0,
	"Class"	INTEGER NOT NULL DEFAULT 0,
	"Phone"	TEXT,
	"Email"	TEXT,
	"NoKName"	TEXT,
	"NoKRelation"	TEXT,
	"NoKPhone"	TEXT,
	"Cohort"	INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY("EntrantID")
);
COMMIT;
