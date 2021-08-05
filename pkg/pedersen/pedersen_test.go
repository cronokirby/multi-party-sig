package pedersen

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/cronokirby/safenum"
	"github.com/taurusgroup/multi-party-sig/pkg/math/sample"
)

var benchParams *Parameters
var benchN *safenum.Modulus

func init() {
	p, _ := new(big.Int).SetString("167787378124493251143019317148640002704627246441948478667228653077480460722587683232574642646881918291617096056059094676028553654483234130554846251885046454551940487805308868634255323683863713140655564735924266284866837842194990483103880904971258203143094403837857522074121191003497549968517847389130694872243", 10)
	q, _ := new(big.Int).SetString("172628447884705665617889981568659580779158791788739623181155628021912172507506304588104455578938133266991401111847751057649396152800487084926965483859584170262709704771256788476402116814674555111681159311762456805404702593970608068725702669729604972720624950347771947966559580483327952522048377101886306484099", 10)
	s, _ := new(big.Int).SetString("25448182756540222866319501898634270977599693736075116251767811398321761734336492575194265315792355808584718336663782672204688106417478368963300473751214591342337476683702712849651809231550640975602518619906861486655178732000388791254150785099820854279346548346528112423491316470201233748237839902669525854602668864974189022599379366745615480940437093919632552402353187455676393853230065384108683249537894167198489453490494046634358648195677341022149808031953931584669015124559681899432749131037714975312907114971572056736764605968515055562566021058441965744442163176054557494286204809259988469497547181346009783327359", 10)
	t, _ := new(big.Int).SetString("8015316201083753856167999987758855473596936298647823494052059806385270716210155885167716563507417164942757686623899459374045961056288024798233083861233712530311639442975277876900958870072552064612806429411065226076877525019560729216371707422331329043485629581836713694830988754061121605813376665575545793435480194071036877604266918397803979335525965949555983633053355236016000661904600828377166630829936488123935666148193019631612745852632950704726062001705644025539422488832473576716710166218032931524855280744450703912325492865050739662260518593977876485434650101403209615051036271024149130698937966662896144492264", 10)
	benchParams = &Parameters{n: new(big.Int).Mul(p, q), s: s, t: t}
	benchN = safenum.ModulusFromNat(new(safenum.Nat).SetBig(benchParams.N(), benchParams.N().BitLen()))
}

// These exist to avoid optimization.
var resultBig *big.Int
var resultBool bool

func BenchmarkPedersenCommit(b *testing.B) {
	b.StopTimer()
	x := sample.IntervalL(rand.Reader)
	y := sample.IntervalL(rand.Reader)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		resultBig = benchParams.Commit(x, y)
	}
}

func BenchmarkPedersenVerify(b *testing.B) {
	b.StopTimer()
	x := sample.IntervalL(rand.Reader).Big()
	y := sample.IntervalL(rand.Reader).Big()
	S := sample.ModN(rand.Reader, benchN).Big()
	T := sample.ModN(rand.Reader, benchN).Big()
	e := sample.IntervalL(rand.Reader).Big()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		resultBool = benchParams.Verify(x, y, S, T, e)
	}
}
