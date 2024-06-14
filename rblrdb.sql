BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "config" (
	"DBInitialised"	INTEGER NOT NULL DEFAULT 1,
	"StartTime"	TEXT NOT NULL DEFAULT '05:00',
	"StartCohortMins"	INTEGER NOT NULL DEFAULT 10,
	"ExtraCohorts"	INTEGER NOT NULL DEFAULT 3
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
	"EntryDonation"	NUMERIC,
	"SquiresCheque"	NUMERIC,
	"SquiresCash"	NUMERIC,
	"RBLRAccount"	NUMERIC,
	"JustGivingAmt"	NUMERIC,
	"JustGivingURL"	TEXT,
	PRIMARY KEY("EntrantID")
);
COMMIT;
