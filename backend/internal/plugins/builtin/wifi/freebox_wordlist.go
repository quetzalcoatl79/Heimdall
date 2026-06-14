package wifi

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FreeboxLatinWords contains the ~800 most common Latin words used by Freebox
// This is a subset of the full 2000+ words used by Free
// Source: Analysis of known Freebox passwords and Latin dictionaries
var FreeboxLatinWords = []string{
	// A
	"abactus", "abalieno", "abavus", "abbas", "abdico", "abdo", "abdomen", "abduco",
	"abeo", "aberro", "abhinc", "abhorreo", "abicio", "abiego", "abigo", "abiuro",
	"ablativus", "abluo", "abnego", "abnuo", "aboleo", "abolitio", "aborior", "abrado",
	"abripio", "abrogo", "abrumpo", "abruptio", "abscedo", "abscido", "abscindo",
	"absconditus", "absens", "absentia", "absisto", "absolvo", "absorbo", "absque",
	"abstinentia", "abstraho", "abstrudo", "absum", "absurdus", "abundans", "abundantia",
	"abundo", "abusio", "abutor", "accedo", "accelero", "accendo", "acceptus", "accido",
	"accipio", "accommodo", "accumulo", "accusator", "accuso", "acer", "acerbus",
	"acervus", "acidus", "acies", "acquiro", "acriter", "actio", "actor", "actuosus",
	"actus", "acumen", "acuo", "acutus", "adamo", "adapto", "adaugeo", "addo", "adduco",
	"ademptio", "adeo", "adeptio", "adequito", "adfero", "adficio", "adfigo", "adfligo",
	// B
	"baca", "bacca", "bacillum", "baculum", "baiulo", "balbus", "balnearius", "balneum",
	"balsamum", "barathrum", "barba", "barbarus", "barbatus", "bardus", "baro", "basio",
	"basilica", "basis", "beatus", "bellator", "bellicus", "bello", "bellum", "bellus",
	"bene", "beneficium", "benevolentia", "benignus", "bestia", "bibo", "biduum", "biennium",
	"bifarius", "bigae", "bini", "bipennis", "biplex", "bis", "bisulcus", "bitumen",
	"blandimentum", "blandior", "blanditia", "blandus", "blasphemus", "bonus", "bos",
	"botrus", "brevis", "brevitas", "breviter", "bruma", "brumalis", "brutus", "bubo",
	// C
	"caballus", "cacumen", "cadaver", "cado", "caducus", "caecitas", "caecus", "caedes",
	"caedo", "caelebs", "caelestis", "caelitus", "caelum", "caenosus", "caenum", "caeremonia",
	"caeruleus", "caesaries", "caesius", "caesar", "caieta", "calamitas", "calamitosus",
	"calamus", "calcar", "calco", "calculus", "caldus", "caleo", "calesco", "caligo",
	"calix", "callidus", "calor", "calumnia", "calvaria", "calvities", "calvus", "calx",
	"camera", "caminus", "campana", "campester", "campus", "canalis", "cancer", "candela",
	"candeo", "candidatus", "candidus", "candor", "canesco", "caninus", "canis", "canistrum",
	"canna", "cano", "canonicus", "canor", "canorus", "canto", "cantus", "canus", "capax",
	"capella", "caper", "capesso", "capillus", "capio", "capitalis", "capitulum", "capto",
	"captus", "caput", "carbo", "carcer", "cardo", "careo", "caries", "cariosus", "caritas",
	"carmen", "carnalis", "carnis", "caro", "carpo", "carus", "casa", "caseus", "cassus",
	"castellum", "castigatio", "castigo", "castra", "castrum", "castus", "casus", "catena",
	"caterva", "cathedra", "catholicus", "cauda", "caupo", "causa", "caute", "cautela",
	"cautio", "cautus", "cavea", "caveo", "caverna", "cavo", "cavus", "cedo", "celebratio",
	"celebro", "celer", "celeritas", "cella", "cellarium", "celo", "celsitudo", "celsus",
	"cena", "cenaculum", "ceno", "censeo", "censor", "censura", "census", "centum",
	"centrum", "centurio", "cera", "cerasus", "cereus", "cerno", "cernuus", "certamen",
	"certe", "certior", "certo", "certus", "cerva", "cervix", "cervus", "cesso", "ceterus",
	// D
	"damno", "damnum", "dapes", "dator", "datum", "debeo", "debilito", "debitor", "decedo",
	"decem", "december", "decens", "decenter", "decerno", "decerto", "decet", "decido",
	"decimatio", "decimus", "decipio", "declamatio", "declaro", "declino", "decor",
	"decoro", "decorus", "decretum", "decumbo", "decurro", "decursus", "decus", "dedecor",
	"dedico", "deduco", "defaeco", "defendo", "defensor", "defero", "defessus", "defetiscor",
	"deficio", "defigo", "definio", "definitio", "defluo", "deformis", "defrudo", "defungor",
	"degenero", "degusto", "deicio", "dein", "deinceps", "deinde", "delectatio", "delecto",
	"delegatio", "delego", "deleo", "delibero", "delibo", "delicatus", "delicia", "delictum",
	"deligo", "delinquo", "delirium", "delitesco", "delubrum", "demens", "dementia",
	"demergo", "demetior", "demigro", "deminuo", "demiror", "demitto", "demo", "demonstro",
	"demoror", "demoveo", "demulceo", "demum", "denarius", "denego", "denique", "denomino",
	"dens", "densus", "denudo", "denuntio", "denuo", "deorsum", "deosculor", "depello",
	"dependo", "depereo", "depleo", "deploro", "depono", "depopulor", "deporto", "deposco",
	"deprecor", "deprehendo", "deprimo", "depromo", "depulso", "deputo", "derelinquo",
	"derideo", "deripio", "desero", "desertum", "desiderium", "desidero", "desidiosus",
	"designo", "desino", "desipio", "desisto", "desolatus", "despecto", "desperatio",
	"despero", "despicio", "despolio", "destino", "destituo", "destruo", "desum", "desumo",
	"desuper", "detego", "deterior", "determino", "detero", "detestor", "detineo", "detraho",
	"detrimentum", "deturbo", "deus", "devasto", "deveho", "devello", "devenio", "devito",
	"devoco", "devolo", "devoror", "devotio", "devoveo", "dexter", "diabolus", "diadema",
	// E
	"ebibo", "ebrius", "ebur", "ecclesia", "econtra", "ecquis", "ecstasis", "edax",
	"edico", "edisco", "editio", "edo", "edoceo", "edomo", "educo", "effectus", "effero",
	"efferus", "efficax", "efficiens", "efficio", "effigia", "efflagito", "effligo",
	"efflo", "efflugo", "effluo", "effluvium", "effor", "effrego", "effugio", "effulgeo",
	"effundo", "effusio", "egelidus", "egens", "egeo", "egero", "egestas", "ego", "egredior",
	"eiectio", "elaboro", "elatus", "electio", "elegans", "elegantia", "elementum", "elenchus",
	"elephantus", "elevo", "elicio", "elido", "eligo", "elimino", "elinguis", "eloquens",
	"eloquentia", "eloquium", "eloquor", "elucesco", "elucido", "elucubro", "eludo", "eluo",
	"eluvies", "emanio", "emano", "emendo", "emergo", "emineo", "eminus", "emissarium",
	"emitto", "emo", "emolumentum", "emorior", "emoveo", "emptor", "emptus", "emulatio",
	"emulator", "emulor", "enarrator", "enarro", "enim", "enitor", "enumero", "enuntio",
	"eo", "ephemeris", "ephippium", "episcopatus", "episcopus", "epistola", "epitome",
	"epulae", "epulor", "epulo", "eques", "equidem", "equitatus", "equus", "erga", "ergo",
	"erigo", "eripio", "erogo", "erro", "error", "erubesco", "erudio", "erumpo", "eruo",
	"erus", "ervum", "escendo", "esse", "essentia", "ester", "esurio", "etiam", "etsi",
	"evacuo", "evado", "evagor", "evalesco", "evanesco", "evanidus", "evangelium", "evasor",
	"eveho", "evello", "evenio", "eventus", "everbero", "everriculum", "eversio", "everto",
	"evidens", "evigilo", "eviscero", "evito", "evoco", "evolo", "evolvo", "evomo", "evulgo",
	// F
	"faber", "fabrica", "fabricor", "fabula", "facies", "facile", "facilis", "facilitas",
	"facillimus", "facio", "factum", "facultas", "facundia", "faenum", "falla", "fallacia",
	"fallaciosus", "fallax", "fallo", "falsidicus", "falso", "falsus", "fama", "fames",
	"familia", "familiaris", "familiaritas", "famosus", "famulus", "fanum", "fari", "farino",
	"fas", "fascino", "fateor", "fatigo", "fatum", "fauces", "fautor", "faveo", "favor",
	"fax", "febris", "fecunditas", "fecundus", "felicitas", "felix", "femina", "femur",
	"fenestra", "fenum", "fera", "ferax", "fere", "feretrum", "ferinus", "ferio", "ferme",
	"fero", "ferox", "ferramentum", "ferraria", "ferratus", "ferreus", "ferrum", "fertilis",
	"fertilitas", "fervens", "ferveo", "fervidus", "fervor", "fessus", "festinatio", "festino",
	"festivus", "festum", "festus", "fetus", "fibre", "fictum", "ficus", "fidelis", "fidelitas",
	"fideliter", "fides", "fiducia", "figo", "figuro", "filia", "filius", "filum", "fimbria",
	"fimus", "findo", "fines", "fingo", "finio", "finis", "finitimus", "finitor", "finitus",
	"fio", "firmamento", "firmamentum", "firmitas", "firmo", "firmus", "fiscus", "fissura",
	"fistula", "fixus", "flabellum", "flaccidus", "flaccus", "flagellum", "flagitium", "flagito",
	"flagrans", "flagro", "flamen", "flamma", "flammeus", "flatus", "flavus", "flebilis", "flecto",
	"fleo", "fletus", "flexibilis", "flexura", "floccus", "floreo", "flos", "fluctuo", "fluctus",
	"fluens", "fluidus", "fluito", "flumen", "fluo", "fluvius", "focus", "fodio", "foedus",
	"foedum", "folium", "folliculus", "fomes", "fons", "fontis", "for", "foras", "forensis",
	"fores", "forfex", "forgo", "foris", "forma", "formalis", "formaliter", "formido",
	"formidolosus", "formo", "formosus", "formula", "fornax", "foro", "fors", "forsitan",
	"fortasse", "forte", "fortis", "fortiter", "fortitudo", "fortuitus", "fortuna", "fortunatus",
	"fortuno", "forum", "forus", "fossa", "fossio", "fossor", "fovea", "foveo", "fractio",
	"fractus", "fragilis", "fragilitas", "fragor", "fragosus", "fragro", "frango", "frater",
	"fraternitas", "fraudatio", "fraudo", "fraus", "fremitus", "fremo", "frendeo", "frendo",
	"frenum", "frequens", "frequentia", "frequento", "fretum", "fretus", "frigeo", "frigesco",
	"frigidus", "frigus", "frixum", "frixus", "frons", "fronten", "frontis", "fructuosus",
	"fructus", "frugi", "frugifer", "frugifero", "fruor", "frustra", "frustratio", "frustro",
	"frustror", "frusto", "frustum", "frux", "fucus", "fuga", "fugax", "fugio", "fugitivus",
	"fugo", "fulcio", "fulgeo", "fulgor", "fulgur", "fulmen", "fultura", "fumidus", "fumo",
	"fumosus", "fumus", "functio", "functus", "fundamentum", "funditor", "funditus", "fundo",
	"fundus", "funebris", "funero", "funestus", "fungor", "funis", "funus", "fur", "furax",
	"furca", "furfur", "furia", "furiosus", "furo", "furor", "furs", "furtificus", "furtim",
	"furtum", "fuscus", "fusio", "futurus",
	// G-H
	"galea", "gaudeo", "gaudium", "gelu", "geminus", "gemma", "gemo", "generalis", "generatim",
	"genero", "genesis", "genetrix", "genialis", "genius", "gens", "gentilis", "genu",
	"genus", "gero", "gestamen", "gestio", "gesto", "gestum", "gilvus", "glacialis", "glacies",
	"gladius", "gleba", "globosus", "globus", "gloria", "glorificatio", "glorifico", "glorior",
	"gloriosus", "gnarus", "gradatim", "gradior", "gradus", "graecia", "graecus", "grammaticus",
	"grandaevus", "grandis", "grando", "grano", "gratanter", "grates", "gratia", "gratis",
	"gratulatio", "gratulor", "gratus", "gravamen", "gravatus", "gravidus", "gravitas",
	"graviter", "gravo", "gravis", "grex", "gubernatio", "guberno", "gula", "gusto", "gustus",
	"guttur", "gutta", "habeo", "habilis", "habitatio", "habito", "habitus", "hactenus",
	"haedus", "haereo", "haeres", "haesito", "harena", "harundo", "hasta", "haud", "haudquaquam",
	"haurio", "haustus", "hebdomas", "hebeo", "hebes", "hebraicus", "hedera", "helluo",
	"hemisphaerium", "herba", "herbosus", "heres", "hereticus", "hermita", "heros", "hesternus",
	"hiatus", "hibernus", "hic", "hiems", "hilaris", "hilaritas", "hinc", "hircus", "hirundo",
	"hispanus", "historia", "hodie", "hodiernus", "holitor", "holocaustum", "homo", "homunculus",
	"honestas", "honesto", "honestus", "honor", "honora", "honorabilis", "hora", "hordeum",
	"horizon", "hornus", "horologium", "horrendus", "horreo", "horresco", "horreum", "horribilis",
	"horridus", "horrifer", "horror", "horsum", "hortatio", "hortator", "hortor", "hortus",
	"hospes", "hospis", "hospitalis", "hospitium", "hostia", "hostilis", "hostis", "huic",
	"humanitas", "humanus", "humecto", "humidus", "humilis", "humilitas", "humo", "humor",
	"humus", "hyacinthus", "hymnus", "hypocrisis", "hypocrita",
	// I-L
	"iaceo", "iacio", "iactantia", "iactatio", "iacto", "iactum", "iactus", "iaculum", "iam",
	"iamdiu", "iamdudum", "iampridem", "ianua", "ibi", "ibidem", "ico", "ictus", "idcirco",
	"idem", "identidem", "ideo", "idoneus", "igitur", "ignarus", "ignavia", "ignavus",
	"igneus", "ignis", "ignobilis", "ignominia", "ignorantia", "ignoro", "ignosco", "ignotus",
	"ilex", "ilico", "illa", "illac", "ille", "illecebra", "illecebrosus", "illi", "illic",
	"illico", "illido", "illigo", "illimis", "illinc", "illino", "illuc", "illuceo", "illucesco",
	"illudo", "illuminatio", "illumino", "illusio", "illustris", "illustro", "illuvies", "imago",
	"imbecillitas", "imbecillus", "imber", "imbibed", "imbrium", "imbuo", "imitatio", "imitator",
	"imitor", "immanitas", "immanis", "immaturus", "immemor", "immensus", "immergo", "immineor",
	"immineo", "imminuo", "immisceo", "immixtio", "immitto", "immo", "immobilis", "immodestus",
	"immolatio", "immolo", "immortalis", "immortalitas", "immotus", "immunis", "immunitas",
	"immuto", "impatientia", "impedimentum", "impedio", "impello", "impendeo", "impendium",
	"impendo", "imperator", "imperatorius", "imperiosus", "imperium", "impero", "impertio",
	"impervius", "impeto", "impetro", "impetus", "impiger", "impingo", "impius", "impleo",
	"implexus", "implicatio", "implico", "imploro", "impono", "importo", "importuna",
	"importunitas", "importunus", "impostor", "impotens", "imprecor", "impressio", "imprimo",
	"improbitas", "improbo", "improbus", "improprius", "improvidus", "improvisum", "imprudens",
	"imprudenter", "imprudentia", "impudens", "impudenter", "impudicus", "impugno", "impulsus",
	"impune", "impunitus", "impurus", "imputatio", "imputo", "imus",
	// M-O
	"machina", "macula", "maculosus", "madefacio", "madesco", "madidus", "maeror", "maestus",
	"magis", "magister", "magistra", "magistratus", "magnifico", "magnificus", "magnitudo",
	"magnus", "maiestas", "maior", "maiores", "male", "maleficio", "maleficus", "maledico",
	"maledicta", "malevolens", "malevolentia", "malignus", "malitia", "malo", "malum",
	"malus", "mancipo", "mandatum", "mando", "mane", "maneo", "manifesto", "manifestus",
	"manipulus", "manna", "mano", "mansuesco", "mansuetudo", "mansuetus", "mantum", "manus",
	"mare", "margo", "maritus", "marmor", "martyrium", "massa", "mater", "materia", "maternus",
	"matrimonium", "matrona", "maturo", "maturus", "maxime", "maximus", "meatus", "mecum",
	"medicina", "medicinus", "medicus", "medietas", "mediocris", "mediocritas", "meditor",
	"medium", "medius", "medulla", "mel", "melior", "membrana", "membrum", "memini", "memor",
	"memorabilis", "memoria", "memoro", "mendacem", "mendacium", "mendax", "mendicitas",
	"mendico", "mendosus", "mens", "mensa", "mensis", "mensura", "mentior", "mentis",
	"mentum", "mercator", "merces", "merda", "mereor", "meretrix", "mergo", "meridianus",
	"meridies", "merito", "meritum", "merus", "messis", "meta", "metallum", "metior",
	"meto", "metuo", "metus", "meus", "mica", "miles", "milia", "militaris", "militia",
	"milito", "mille", "minae", "minax", "minime", "minimus", "ministerium", "ministro",
	"minor", "minuo", "minus", "mirabilis", "miraculum", "miror", "mirus", "misceo",
	"misellus", "miser", "miserabilis", "miseratio", "misere", "misereo", "misereor",
	"miseria", "misericordia", "missa", "mitigo", "mitis", "mitto", "mobilis", "mobilitas",
	"moderamen", "moderatio", "moderator", "moderor", "modestia", "modestus", "modicus",
	"modifico", "modo", "modus", "moenia", "molestia", "molesto", "molestus", "molimen",
	"molior", "mollis", "mollitia", "momentum", "monachus", "monarchia", "monasterium",
	"moneo", "monimentum", "monitio", "monitor", "monitus", "mons", "monstrum", "monstro",
	"monumentum", "mora", "moralis", "morbus", "mordax", "mordeo", "moribundus", "morior",
	"moror", "mors", "morsum", "morsus", "mortalis", "mortalitas", "mortifer", "mortifico",
	"mortis", "mortuus", "morum", "morus", "mos", "motio", "moto", "motus", "moveo",
	"mox", "mucro", "mugitus", "mulceo", "mulier", "muliercula", "multiformis", "multifrons",
	"multimode", "multiplex", "multiplicatio", "multiplico", "multitudo", "multus", "multo",
	"mundanus", "munde", "munditia", "mundo", "mundus", "munero", "muneratus", "munifice",
	"munificentia", "munificus", "munimen", "munimentum", "munio", "munitio", "munus",
	"murex", "murmur", "murmuro", "murus", "mus", "musca", "musculus", "mussito", "muto",
	"mutatio", "muto", "mutuor", "mutuo", "mutuus", "mysterium",
	// N-P
	"nam", "namque", "narro", "nascor", "natio", "nativitas", "nato", "natura", "naturalis",
	"natus", "naufragus", "nauta", "navigo", "navis", "necdum", "necessarius", "necesse",
	"necessitas", "neco", "nefas", "nefastus", "negatio", "neglego", "negotiator", "negotium",
	"nemo", "nemus", "neo", "nepos", "nequam", "nequaquam", "neque", "nequeo", "nequior",
	"nequiter", "nervus", "nescio", "nescius", "neuter", "neutiquam", "neve", "nexum", "nexus",
	"nihil", "nihilominus", "nihilum", "nimio", "nimis", "nimium", "nimius", "nisi", "nisus",
	"niteo", "nitesco", "nitidus", "nitor", "niveus", "nix", "nobilis", "nobilitas", "nobilito",
	"noceo", "noctis", "nocturnus", "nodus", "nolo", "nomen", "nominatim", "nomino", "non",
	"nondum", "nonnisi", "nonnullus", "nonnumquam", "nonus", "norma", "nosco", "noster",
	"nota", "notabilis", "notarius", "notitia", "noto", "notus", "novem", "novitas", "novo",
	"novus", "nox", "noxa", "noxia", "nubes", "nubila", "nubo", "nucleus", "nudus", "nullus",
	"numen", "numerosus", "numero", "numerus", "nummus", "numquam", "nunc", "nuncupo",
	"nundinae", "nuntio", "nuntius", "nuo", "nuper", "nuptiae", "nurus", "nusquam", "nutrio",
	"nutritus", "nutus", "nux",
	// O-P (partial)
	"obaeratus", "obdormio", "obduco", "obduro", "obeo", "obfirmatus", "obfirmo", "obiectus",
	"obiicio", "obitus", "obiurgo", "oblatio", "oblecto", "obligo", "obliquus", "oblitteratus",
	"oblitus", "oblivio", "obliviscor", "obloquor", "obnoxius", "oboedio", "obreptio", "obrepo",
	"obruo", "obscenus", "obscurum", "obscurus", "obsequens", "obsequium", "obsequor", "observantia",
	"observatio", "observo", "obses", "obsessio", "obsideo", "obsidio", "obsolesco", "obstaculum",
	"obstinatio", "obstinatus", "obstino", "obsto", "obstruo", "obstupefacio", "obsum", "obtemperatio",
	"obtempero", "obtineo", "obtingo", "obtrudo", "obtutus", "obumbro", "obviam", "obvius",
	"occasio", "occasus", "occido", "occisor", "occulto", "occultus", "occupatio", "occupo",
	"occurro", "occursus", "oceanus", "ocellus", "ocior", "octo", "oculus", "odio", "odiosus",
	"odium", "odor", "odoratus", "offendo", "offensa", "offensio", "offero", "officina",
	"officiosus", "officium", "olim", "oliva", "olla", "omen", "ominor", "omitto", "omnifariam",
	"omnino", "omnipotens", "omnis", "oneratus", "onero", "onus", "opacus", "opera", "operatio",
	"operio", "operor", "opes", "opifex", "opimus", "opinio", "opinor", "opobalsamum", "oportet",
	"oportuno", "opperior", "oppeto", "oppido", "oppidum", "oppono", "opportunitas", "opportune",
	"opportunus", "oppressio", "opprimo", "opprobrium", "oppugno", "ops", "optatus", "optimates",
	"optimus", "opto", "opulens", "opulentia", "opus", "ora", "oraculum", "oratio", "orator",
	"oratorius", "orbis", "orbitas", "orbo", "orbus", "ordino", "ordo", "oriens", "origo",
	"orior", "oriundus", "ornamentum", "ornatus", "orno", "oro", "orphanus", "ortus", "os",
	"osculum", "ostendo", "ostento", "ostentum", "ostium", "otiosus", "otium", "ovis", "ovum",
	// P
	"pabulor", "pabulum", "pacatus", "paco", "pactio", "pactum", "paeniteo", "paenitet",
	"paene", "paganus", "pagina", "pala", "palam", "palatium", "palea", "pallium", "palma",
	"palpo", "palus", "panis", "pannus", "pango", "papa", "papaver", "par", "parabola",
	"paradisus", "paratus", "parce", "parco", "parcus", "parens", "pareo", "paries", "pario",
	"pariter", "paro", "parochia", "pars", "parsimonia", "particeps", "participatio", "participo",
	"particula", "partim", "partior", "partis", "parum", "parumper", "parvulus", "parvus",
	"pasco", "pascuus", "passer", "passim", "passio", "passivus", "passus", "pastor", "patefacio",
	"pateo", "pater", "paternoster", "paternus", "patesco", "patientia", "patiens", "patior",
	"patria", "patriarcha", "patrimonium", "patrius", "patrocinor", "patronus", "patrum",
	"patruus", "pauci", "paulatim", "paulisper", "paulo", "pauper", "paupertas", "pavor",
	"pax", "peccamen", "peccator", "peccatum", "pecco", "pecto", "pectus", "peculiaris",
	"peculium", "pecunia", "pecus", "peior", "pello", "pendeo", "pendo", "penes", "penetralis",
	"penetro", "penitens", "penna", "penso", "pensum", "penuria", "penus", "pepo", "per",
	"peracto", "peragito", "perago", "peramplus", "perattente", "percello", "percipio", "percutio",
	"perdo", "perduco", "pereo", "peregrinus", "perennis", "perfectio", "perfectus", "perfero",
	"perficio", "perfidia", "perfidiosus", "perfidus", "perfringo", "perfruor", "perfugio",
	"perfundo", "pergo", "periclitor", "periculum", "perimo", "peritus", "periurium", "periurus",
	"perlustro", "permagnificus", "permaneo", "permitto", "permoveo", "perpetro", "perpetuus",
	"perplexa", "perscribo", "perscrutor", "persequor", "persevero", "persolvo", "persona",
	"perspicax", "perspicio", "perspicuus", "persto", "persuadeo", "perterreo", "pertinacia",
	"pertinax", "pertineo", "pertorqueo", "pertraho", "perturbatio", "perturbo", "peruro",
	"pervagor", "pervenio", "perverto", "pervideo", "pervigilo", "pes", "pessimus", "pestis",
	"petitus", "peto", "petulans", "phasma", "philosophus", "physicus", "picea", "pietas",
	"piger", "pignus", "pila", "pilum", "pilus", "pinguis", "pinna", "pius", "placeo",
	"placidus", "placitum", "placo", "plaga", "plango", "planities", "planto", "planus",
	"plasmator", "platea", "plaustrum", "plebs", "plecto", "plene", "plenitudo", "plenus",
	"plerumque", "plerusque", "plico", "ploro", "pluit", "pluma", "plumbeus", "plumbum",
	"pluo", "plures", "plurimi", "plurimum", "plurimus", "plus", "poculum", "poema", "poena",
	"poenalis", "poenitet", "poeta", "polleo", "polliceor", "polluo", "pompa", "pondero",
	"pondus", "pono", "pons", "pontificalis", "pontifex", "pontus", "populus", "porro",
	"porta", "portendo", "portentum", "porto", "portus", "posco", "positio", "positus",
	"possessio", "possideo", "possum", "post", "postea", "posteri", "posteritas", "posterus",
	"posthac", "postis", "postmodum", "postquam", "postremo", "postremus", "postulatio",
	"postulo", "potens", "potentia", "potestas", "potior", "potissimum", "potius", "poto",
	"potus", "prae", "praebeo", "praecedo", "praeceptio", "praeceptor", "praeceptum", "praecido",
	"praecipio", "praecipue", "praeclarus", "praeco", "praeda", "praedator", "praedecessor",
	"praedico", "praeditus", "praedo", "praefero", "praeficio", "praefigo", "praelatus",
	"praemitto", "praemium", "praemo", "praenuntio", "praeparo", "praepono", "praepositus",
	"praerogativa", "praesagio", "praesago", "praescius", "praesens", "praesentia", "praesentium",
	"praesertim", "praeses", "praesidium", "praesieo", "praestans", "praestantia", "praesto",
	"praestruo", "praesum", "praeter", "praeterea", "praetereo", "praeteritus", "praetermitto",
	"praetexta", "praetor", "praetorium", "praevaleo", "praevenio", "praevenio", "pravitas",
	"pravus", "precatus", "preces", "precor", "prehendo", "prelum", "premo", "presbyter",
	"pretiosus", "pretium", "prex", "pridem", "pridie", "primas", "primigenius", "primitus",
	"primo", "primordium", "primoris", "primum", "primus", "princeps", "principalis", "principatus",
	"principio", "principium", "prior", "priscus", "pristinus", "prius", "priusquam", "privatim",
	"privatus", "privo", "pro", "proavus", "probatio", "probitas", "probo", "probus", "procedo",
	"procella", "procer", "procerus", "processio", "processus", "procido", "proclamo", "procrastino",
	"procreo", "procul", "procurator", "procuro", "prodeo", "prodico", "prodigium", "prodigo",
	"prodigus", "proditio", "proditor", "prodo", "prodromus", "produco", "proelium", "profanus",
	"profectio", "profecto", "profectus", "profero", "professio", "professor", "proficio",
	"proficiscor", "profiteor", "profligo", "profluo", "profugio", "profugus", "profundo",
	"profundus", "progenero", "progenies", "progenitor", "progigno", "progredior", "progressio",
	"progressus", "prohibeo", "proicio", "proinde", "prolemma", "proles", "prolixus", "proloquium",
	"proloquor", "proluo", "promereo", "promereor", "promineo", "promiscuus", "promissio",
	"promissor", "promitto", "promo", "promontorium", "promoveo", "prompte", "promptu", "promptus",
	"promulgo", "promus", "pronuntio", "propagatio", "propago", "prope", "propello", "propemodum",
	"propero", "propinquo", "propinquus", "propior", "propitio", "propitius", "propono",
	"proportio", "propositum", "proprie", "proprietas", "proprius", "propter", "propterea",
	"propugnaculum", "propugno", "prorsum", "prorsus", "prosequor", "prosilio", "prospecto",
	"prospectus", "prosper", "prosperitas", "prospicio", "prosterno", "prostituo", "protego",
	"protendo", "protervus", "protinus", "protraho", "protrudo", "proturbo", "prout", "proveho",
	"provenio", "proventus", "provideo", "providentia", "providus", "provincia", "provocatio",
	"provoco", "provolutus", "provolvo", "proximus", "prudens", "prudenter", "prudentia",
	"pruina", "pruna", "prurio", "publice", "publico", "publicus", "pudens", "pudeo", "pudicus",
	"pudor", "puella", "puer", "puerilis", "pugil", "pugna", "pugnaculum", "pugno", "pugnus",
	"pulcher", "pulchre", "pulchritudo", "pullulo", "pullus", "pulmentum", "pulmo", "pulpa",
	"pulsus", "pulverulentus", "pulvis", "pumex", "pumilius", "punctum", "pungo", "punio",
	"punitio", "punitor", "pupo", "pupillus", "puppis", "purgo", "puritas", "puro", "purpura",
	"purpureus", "purus", "pusillus", "putamen", "putator", "puteo", "puto", "putredo", "putridus",
}

// FreeboxSuffixes contains common suffixes added to Freebox passwords
var FreeboxSuffixes = []string{
	"", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"%", "#", "*", "%%", "##", "**", "%#", "#%", "*%", "%*", "#*", "*#",
	"0%", "1%", "2%", "3%", "4%", "5%", "6%", "7%", "8%", "9%",
	"0#", "1#", "2#", "3#", "4#", "5#", "6#", "7#", "8#", "9#",
	"0*", "1*", "2*", "3*", "4*", "5*", "6*", "7*", "8*", "9*",
}

// GenerateFreeboxWordlist creates a wordlist file for Freebox passwords
// It generates combinations of 4 Latin words with possible suffixes
// Returns the path to the generated wordlist
func GenerateFreeboxWordlist(outputDir string, maxCombinations int) (string, error) {
	outputPath := filepath.Join(outputDir, "freebox-latin-words.txt")

	// First, create a simple wordlist of just the Latin words (for quick dictionary attack)
	simpleWordsPath := filepath.Join(outputDir, "latin-words-simple.txt")
	simpleFile, err := os.Create(simpleWordsPath)
	if err != nil {
		return "", fmt.Errorf("failed to create simple words file: %v", err)
	}

	for _, word := range FreeboxLatinWords {
		fmt.Fprintln(simpleFile, word)
	}
	simpleFile.Close()

	// Now create combination file with pattern: word1-word2-word3-word4
	file, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create wordlist file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	count := 0
	numWords := len(FreeboxLatinWords)
	numSuffixes := len(FreeboxSuffixes)

	// Generate combinations - word1-word2suffix-word3suffix-word4suffix
	// First word usually has no suffix, others may have
	for i := 0; i < numWords && count < maxCombinations; i++ {
		word1 := FreeboxLatinWords[i]
		for j := 0; j < numWords && count < maxCombinations; j++ {
			word2 := FreeboxLatinWords[j]
			// Apply common suffixes to word2
			for s2 := 0; s2 < numSuffixes && count < maxCombinations; s2++ {
				word2Full := word2 + FreeboxSuffixes[s2]

				for k := 0; k < numWords && count < maxCombinations; k++ {
					word3 := FreeboxLatinWords[k]
					for s3 := 0; s3 < numSuffixes && count < maxCombinations; s3++ {
						word3Full := word3 + FreeboxSuffixes[s3]

						for l := 0; l < numWords && count < maxCombinations; l++ {
							word4 := FreeboxLatinWords[l]
							for s4 := 0; s4 < numSuffixes && count < maxCombinations; s4++ {
								word4Full := word4 + FreeboxSuffixes[s4]

								password := fmt.Sprintf("%s-%s-%s-%s", word1, word2Full, word3Full, word4Full)
								fmt.Fprintln(writer, password)
								count++

								if count >= maxCombinations {
									break
								}
							}
						}
					}
				}
			}
		}
	}

	return outputPath, nil
}

// GenerateFreeboxWordlistFast creates a smaller, prioritized wordlist
// focusing on the most common word patterns
func GenerateFreeboxWordlistFast(outputDir string) (string, error) {
	outputPath := filepath.Join(outputDir, "freebox-priority.txt")

	file, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create wordlist file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Use only most common words (first 200)
	priorityWords := FreeboxLatinWords[:min(200, len(FreeboxLatinWords))]
	// Use only common suffixes
	prioritySuffixes := []string{"", "0", "1", "2", "%", "#", "*", "1%", "2%", "1#", "2#"}

	count := 0
	for _, w1 := range priorityWords {
		for _, w2 := range priorityWords {
			for _, s2 := range prioritySuffixes {
				for _, w3 := range priorityWords {
					for _, s3 := range prioritySuffixes {
						for _, w4 := range priorityWords {
							for _, s4 := range prioritySuffixes {
								password := fmt.Sprintf("%s-%s%s-%s%s-%s%s",
									w1, w2, s2, w3, s3, w4, s4)
								fmt.Fprintln(writer, password)
								count++
							}
						}
					}
				}
			}
		}
	}

	return outputPath, nil
}

// GetFreeboxWordlistPath returns the path to the Freebox wordlist, creating it if needed
func GetFreeboxWordlistPath() string {
	wordlistDir := "/opt/heimdall/wordlists"
	wordlistPath := filepath.Join(wordlistDir, "latin-words-simple.txt")

	// Check if exists
	if _, err := os.Stat(wordlistPath); err == nil {
		return wordlistPath
	}

	// Create directory if needed
	os.MkdirAll(wordlistDir, 0755)

	// Create simple words file
	file, err := os.Create(wordlistPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	for _, word := range FreeboxLatinWords {
		fmt.Fprintln(file, word)
	}

	return wordlistPath
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SanitizeFreeboxSSID extracts useful info from a Freebox SSID
func SanitizeFreeboxSSID(ssid string) (model string, suffix string) {
	ssid = strings.ToLower(ssid)
	if strings.HasPrefix(ssid, "freebox-") {
		suffix = strings.TrimPrefix(ssid, "freebox-")
		return "freebox", suffix
	}
	if strings.HasPrefix(ssid, "free_") {
		suffix = strings.TrimPrefix(ssid, "free_")
		return "free", suffix
	}
	return "", ""
}
