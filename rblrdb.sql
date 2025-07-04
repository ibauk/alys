BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "config" (
	"DBInitialised"	INTEGER NOT NULL DEFAULT 1,
	"StartTime"	TEXT NOT NULL DEFAULT '05:00',
	"StartCohortMins"	INTEGER NOT NULL DEFAULT 10,
	"ExtraCohorts"	INTEGER NOT NULL DEFAULT 3,
	"RallyStatus"	TEXT NOT NULL DEFAULT 'S',
	"MinDonation"	INTEGER NOT NULL DEFAULT 50,
	"JustGCharities"	TEXT NOT NULL DEFAULT 219279,
	"MinOdoDiff"	INTEGER NOT NULL DEFAULT 500,
	"MaxOdoDiff"	INTEGER NOT NULL DEFAULT 2000
);
INSERT INTO config(DBInitialised) VALUES(1);
CREATE TABLE IF NOT EXISTS "entrants" (
	"EntrantID"	INTEGER NOT NULL,
	"Bike"	TEXT DEFAULT 'motorbike',
	"BikeReg"	INTEGER,
	"RiderFirst"	TEXT,
	"RiderLast"	TEXT,
	"RiderAddress1"	TEXT,
	"RiderAddress2"	TEXT,
	"RiderTown"	TEXT,
	"RiderCounty"	TEXT,
	"RiderPostcode"	TEXT,
	"RiderCountry"	TEXT DEFAULT 'United Kingdom',
	"RiderIBA"	TEXT,
	"RiderRBL"	TEXT,
	"RiderPhone"	TEXT,
	"RiderEmail"	TEXT,
	"PillionFirst"	TEXT,
	"PillionLast"	TEXT,
	"PillionAddress1"	TEXT,
	"PillionAddress2"	TEXT,
	"PillionTown"	TEXT,
	"PillionCounty"	TEXT,
	"PillionPostcode"	TEXT,
	"PillionCountry"	INTEGER DEFAULT 'United Kingdom',
	"PillionIBA"	TEXT,
	"PillionRBL"	TEXT,
	"OdoCounts"	TEXT NOT NULL DEFAULT 'M',
	"OdoStart"	INTEGER,
	"OdoFinish"	INTEGER,
	"CorrectedMiles"	TEXT DEFAULT 0,
	"FinishTime"	TEXT,
	"StartTime"	TEXT,
	"EntrantStatus"	INTEGER NOT NULL DEFAULT 0,
	"NoKName"	TEXT,
	"NoKRelation"	TEXT,
	"NoKPhone"	TEXT,
	"EntryDonation"	TEXT,
	"SquiresCheque"	TEXT,
	"SquiresCash"	TEXT,
	"RBLRAccount"	TEXT,
	"JustGivingAmt"	TEXT,
	"JustGivingURL"	TEXT,
	"Route"	TEXT DEFAULT 'A-NCW',
	"PillionEmail"	TEXT,
	"PillionPhone"	TEXT,
	"RiderRBLR"	TEXT,
	"PillionRBLR"	TEXT,
	"Tshirt1"	TEXT DEFAULT 'no thanks',
	"Tshirt2"	TEXT DEFAULT 'no thanks',
	"Patches"	INTEGER DEFAULT 0,
	"FreeCamping"	TEXT NOT NULL DEFAULT 'N',
	"CertificateDelivered"	TEXT NOT NULL DEFAULT 'N',
	"CertificateAvailable"	TEXT NOT NULL DEFAULT 'N',
	"Notes"	TEXT,
	"JustGivingPSN"	TEXT,
	"Verified"	TEXT NOT NULL DEFAULT 'N',
	"CertificateStatus"	TEXT NOT NULL DEFAULT 'N',
	PRIMARY KEY("EntrantID")
);
CREATE TABLE IF NOT EXISTS "justgs" (
	"PageShortName"	TEXT NOT NULL,
	"NumUsers"	INTEGER NOT NULL DEFAULT 0,
	"FundsRaised"	INTEGER NOT NULL DEFAULT 0,
	"PerUser"	INTEGER NOT NULL DEFAULT 0,
	"PageValid"	INTEGER NOT NULL DEFAULT 1,
	"CharityReg"	TEXT NOT NULL DEFAULT '',
	"CharityName"	TEXT NOT NULL DEFAULT '',
	PRIMARY KEY("PageShortName")
);
CREATE TABLE IF NOT EXISTS "rallyparams" (
	"RallyTitle"	TEXT,
	"StartTime"	TEXT,
	"FinishTime"	TEXT,
	"DBVersion"	INTEGER DEFAULT 21
);
INSERT INTO rallyparams (RallyTitle,StartTime,FinishTime) VALUES('RBLR1000','2026-06-06T05:00','2026-06-07T12:00');
COMMIT;
