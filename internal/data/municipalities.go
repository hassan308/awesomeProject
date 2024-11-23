package data

import (
	"log"
	"strings"
	"unicode"
)

// MunicipalityMap innehåller mappning mellan kommunnamn och deras ID
var MunicipalityMap = map[string]string{
	"Blekinge län":     "DQZd_uYs_oKb",
	"Karlshamn":        "HtGW_WgR_dpE",
	"Karlskrona":       "YSt4_bAa_ccs",
	"Olofström":        "1gEC_kvM_TXK",
	"Ronneby":          "vH8x_gVz_z7R",
	"Sölvesborg":       "EVPy_phD_8Vf",
	"Dalarnas län":     "oDpK_oZ2_WYt",
	"Avesta":           "Szbq_2fg_ydQ",
	"Borlänge":         "cpya_jJg_pGp",
	"Falun":            "N1wJ_Cuu_7Cs",
	"Gagnef":           "Nn7p_W3Z_y68",
	"Hedemora":         "DE9u_V4K_Z1S",
	"Leksand":          "7Zsu_ant_gcn",
	"Ludvika":          "Ny2b_2bo_7EL",
	"Malung-Sälen":     "FPCd_poj_3tq",
	"Mora":             "UGcC_kYx_fTs",
	"Orsa":             "CRyF_5Jg_4ht",
	"Rättvik":          "Jy3D_2ux_dg8",
	"Smedjebacken":     "5zZX_8FH_Sbq",
	"Säter":            "c3Zx_jBf_CqF",
	"Vansbro":          "4eS9_HX1_M7V",
	"Älvdalen":         "cZtt_qGo_oBr",
	"Gotlands län":     "K8iD_VQv_2BA",
	"Gotland":          "Ft9P_E8F_VLJ",
	"Gävleborgs län":   "zupA_8Nt_xcD",
	"Bollnäs":          "KxjG_ig5_exF",
	"Gävle":            "qk8Y_2b6_82D",
	"Hofors":           "yuNd_3bg_ttc",
	"Hudiksvall":       "Utks_mwF_axY",
	"Ljusdal":          "63iQ_V6F_REB",
	"Nordanstig":       "fFeF_RCz_Tm5",
	"Ockelbo":          "GEvW_wKy_A9H",
	"Ovanåker":         "JPSe_mUQ_NDs",
	"Sandviken":        "BbdN_xLB_k6s",
	"Söderhamn":        "JauG_nz5_7mu",
	"Hallands län":     "wjee_qH2_yb6",
	"Falkenberg":       "qaJg_wMR_C8T",
	"Halmstad":         "kUQB_KdK_kAh",
	"Hylte":            "3XMe_nGt_RcU",
	"Kungsbacka":       "3JKV_KSK_x6z",
	"Laholm":           "c1iL_rqh_Zja",
	"Varberg":          "AkUx_yAq_kGr",
	"Jämtlands län":    "65Ms_7r1_RTG",
	"Berg":             "gRNJ_hVW_Gpg",
	"Bräcke":           "eNSc_Nj1_CDP",
	"Härjedalen":       "j35Q_VKL_NiM",
	"Krokom":           "yurW_aLE_4ga",
	"Ragunda":          "Voto_egJ_FZP",
	"Strömsund":        "ppjq_Eci_Wz9",
	"Åre":              "D7ax_CXP_6r1",
	"Östersund":        "Vt7P_856_WZS",
	"Jönköpings län":   "MtbE_xWT_eMi",
	"Aneby":            "y9HE_XD7_WaD",
	"Eksjö":            "VacK_WF6_XVg",
	"Gislaved":         "cNQx_Yqi_83Q",
	"Gnosjö":           "91VR_Hxi_GN4",
	"Habo":             "9zQB_3vU_BQA",
	"Jönköping":        "KURg_KJF_Lwc",
	"Mullsjö":          "smXg_BXp_jTW",
	"Nässjö":           "KfXT_ySA_do2",
	"Sävsjö":           "L1cX_MjM_y8W",
	"Tranås":           "Namm_SpC_RPG",
	"Vaggeryd":         "zFup_umX_LVv",
	"Vetlanda":         "xJqx_SLC_415",
	"Värnamo":          "6bS8_fzf_xpW",
	"Kalmar län":       "9QUH_2bb_6Np",
	"Borgholm":         "LY9i_qNL_kXf",
	"Emmaboda":         "1koj_6Bg_8K6",
	"Hultsfred":        "AEQD_1RT_vM9",
	"Högsby":           "WPDh_pMr_RLZ",
	"Kalmar":           "Pnmg_SgP_uHQ",
	"Mönsterås":        "8eEp_iz4_cNN",
	"Mörbylånga":       "Muim_EAi_EFp",
	"Nybro":            "xk68_bJa_6Fh",
	"Oskarshamn":       "tUP8_hRE_NcF",
	"Torsås":           "wYFb_q7w_Nnh",
	"Vimmerby":         "a7hJ_zwv_2FR",
	"Västervik":        "t7H4_S2P_3Fw",
	"Kronobergs län":   "tF3y_MF9_h5G",
	"Alvesta":          "MMph_wmN_esc",
	"Lessebo":          "nXZy_1Jd_D8X",
	"Ljungby":          "GzKo_S48_QCm",
	"Markaryd":         "ZhVf_yL5_Q5g",
	"Tingsryd":         "qz8Q_kDz_N2Y",
	"Uppvidinge":       "78cu_S5T_Pgp",
	"Växjö":            "mmot_H3A_auW",
	"Älmhult":          "EK6X_wZq_CQ8",
	"Norrbottens län":  "9hXe_F4g_eTG",
	"Arjeplog":         "vkQW_GB6_MNk",
	"Arvidsjaur":       "A5WX_XVo_Zt6",
	"Boden":            "y4NQ_tnB_eVd",
	"Gällivare":        "6R2u_zkb_uoS",
	"Haparanda":        "tfRE_hXa_eq7",
	"Jokkmokk":         "mp6j_2b6_1bz",
	"Kalix":            "cUyN_C9V_HLU",
	"Kiruna":           "biN6_UiL_Qob",
	"Luleå":            "CXbY_gui_14v",
	"Pajala":           "dHMF_72G_4NM",
	"Piteå":            "umej_bP2_PpK",
	"Älvsbyn":          "14WF_zh1_W3y",
	"Överkalix":        "n5Sq_xxo_QWL",
	"Övertorneå":       "ehMP_onv_Chk",
	"Skåne län":        "CaRE_1nn_cSU",
	"Bjuv":             "waQp_FjW_qhF",
	"Bromölla":         "WMNK_PXa_Khm",
	"Burlöv":           "64g5_Lio_aMU",
	"Båstad":           "i8vK_odq_6ar",
	"Eslöv":            "gfCw_egj_1M4",
	"Helsingborg":      "qj3q_oXH_MGR",
	"Hässleholm":       "bP5q_53x_aqJ",
	"Höganäs":          "8QQ6_e95_a1d",
	"Hörby":            "autr_KMa_cfp",
	"Höör":             "N29z_AqQ_Ppc",
	"Klippan":          "JARU_FAY_hTS",
	"Kristianstad":     "vrvW_sr8_1en",
	"Kävlinge":         "5ohg_WJU_Ktn",
	"Landskrona":       "Yt5s_Vf9_rds",
	"Lomma":            "naG4_AUS_z2v",
	"Lund":             "muSY_tsR_vDZ",
	"Malmö":            "oYPt_yRA_Smm",
	"Osby":             "najS_Lvy_mDD",
	"Perstorp":         "BN7r_iPV_F9p",
	"Simrishamn":       "dLxo_EpC_oPe",
	"Sjöbo":            "P3Cs_1ZP_9XB",
	"Skurup":           "oezL_78x_r89",
	"Staffanstorp":     "vBrj_bov_KEX",
	"Svalöv":           "2r6J_g2w_qp5",
	"Svedala":          "n6r4_fjK_kRr",
	"Tomelilla":        "UMev_wGs_9bg",
	"Trelleborg":       "STvk_dra_M1X",
	"Vellinge":         "Tcog_5sH_b46",
	"Ystad":            "hdYk_hnP_uju",
	"Ängelholm":        "pCuv_P5A_9oh",
	"Åstorp":           "tEv6_ktG_QQb",
	"Örkelljunga":      "nBTS_Nge_dVH",
	"Östra Göinge":     "LTt7_CGG_RUf",
	"Stockholms län":   "CifL_Rzy_Mku",
	"Botkyrka":         "CCVZ_JA7_d3y",
	"Danderyd":         "E4CV_a5E_ucX",
	"Ekerö":            "magF_Gon_YL2",
	"Haninge":          "Q7gp_9dT_k2F",
	"Huddinge":         "g1Gc_aXK_EKu",
	"Järfälla":         "qm5H_jsD_fUF",
	"Lidingö":          "FBbF_mda_TYD",
	"Nacka":            "aYA7_PpG_BqP",
	"Norrtälje":        "btgf_fS7_sKG",
	"Nykvarn":          "mBKv_q3B_SK8",
	"Nynäshamn":        "37UU_T7x_oxG",
	"Salem":            "4KBw_CPU_VQv",
	"Sigtuna":          "8ryy_X54_xJj",
	"Sollentuna":       "Z5Cq_SgB_dsB",
	"Solna":            "zHxw_uJZ_NJ8",
	"Stockholm":        "AvNB_uwa_6n6",
	"Sundbyberg":       "UTJZ_zHH_mJm",
	"Södertälje":       "g6hK_M1o_hiU",
	"Tyresö":           "sTPc_k2B_SqV",
	"Täby":             "onpA_B5a_zfv",
	"Upplands Väsby":   "XWKY_c49_5nv",
	"Upplands-Bro":     "w6yq_CGR_Fiv",
	"Vallentuna":       "K4az_Bm6_hRV",
	"Vaxholm":          "9aAJ_j6L_DST",
	"Värmdö":           "15nx_Vut_GrH",
	"Österåker":        "8gKt_ZsV_PGj",
	"Södermanlands län": "s93u_BEb_sx2",
	"Eskilstuna":       "kMxr_NiX_YrU",
	"Flen":             "P8yp_WT9_Bks",
	"Gnesta":           "os8Y_RUo_U3u",
	"Katrineholm":      "snx9_qVD_Dr1",
	"Nyköping":         "KzvD_ePV_DKQ",
	"Oxelösund":        "72XK_mUU_CAH",
	"Strängnäs":        "shnD_RiE_RKL",
	"Trosa":            "rjzu_nQn_mCK",
	"Vingåker":         "rut9_f5W_kTX",
	"Uppsala län":      "zBon_eET_fFU",
	"Enköping":         "HGwg_unG_TsG",
	"Heby":             "sD2e_1Tr_4WZ",
	"Håbo":             "Bbs5_JUs_Qh5",
	"Knivsta":          "KALq_sG6_VrW",
	"Tierp":            "K8A2_JBa_e6e",
	"Uppsala":          "otaF_bQY_4ZD",
	"Älvkarleby":       "cbyw_9aK_Cni",
	"Östhammar":        "VE3L_3Ei_XbG",
	"Värmlands län":    "EVVp_h6U_GSZ",
	"Arvika":           "yGue_F32_wev",
	"Eda":              "N5HQ_hfp_7Rm",
	"Filipstad":        "UXir_vKD_FuW",
	"Forshaga":         "xnEt_JN3_GkA",
	"Grums":            "PSNt_P95_x6q",
	"Hagfors":          "qk9a_g5U_sAH",
	"Hammarö":          "x5qW_BXr_aut",
	"Karlstad":         "hRDj_PoV_sFU",
	"Kil":              "ocMw_Rz5_B1L",
	"Kristinehamn":     "SVQS_uwJ_m2B",
	"Munkfors":         "x73h_7rW_mXN",
	"Storfors":         "mPt5_3QD_LTM",
	"Sunne":            "oqNH_cnJ_Tdi",
	"Säffle":           "wmxQ_Guc_dsy",
	"Torsby":           "hQdb_zn9_Sok",
	"Årjäng":           "ymBu_aFc_QJA",
	"Västerbottens län": "g5Tt_CAV_zBd",
	"Bjurholm":         "vQkf_tw2_CmR",
	"Dorotea":          "tSkf_Tbn_rHk",
	"Lycksele":         "7rpN_naz_3Uz",
	"Malå":             "7sHJ_YCE_5Zv",
	"Nordmaling":       "wMab_4Zs_wpM",
	"Norsjö":           "XmpG_vPQ_K7T",
	"Robertsfors":      "p8Mv_377_bxp",
	"Skellefteå":       "kicB_LgH_2Dk",
	"Sorsele":          "VM7L_yJK_Doo",
	"Storuman":         "gQgT_BAk_fMu",
	"Umeå":             "QiGt_BLu_amP",
	"Vilhelmina":       "tUnW_mFo_Hvi",
	"Vindeln":          "izT6_zWu_tta",
	"Vännäs":           "utQc_6xq_Dfm",
	"Åsele":            "xLdL_tMA_JJv",
	"Västernorrlands län": "NvUF_SP1_1zo",
	"Härnösand":        "uYRx_AdM_r4A",
	"Kramfors":         "yR8g_7Jz_HBZ",
	"Sollefteå":        "v5y4_YPe_TMZ",
	"Sundsvall":        "dJbx_FWY_tK6",
	"Timrå":            "oJ8D_rq6_kjt",
	"Ånge":             "swVa_cyS_EMN",
	"Örnsköldsvik":     "zBmE_n6s_MnQ",
	"Västmanlands län": "G6DV_fKE_Viz",
	"Arboga":           "Jkyb_5MQ_7pB",
	"Fagersta":         "7D9G_yrX_AGJ",
	"Hallstahammar":    "oXYf_HmD_ddE",
	"Kungsör":          "Fac5_h7a_UoM",
	"Köping":           "4Taz_AuG_tSm",
	"Norberg":          "jbVe_Cps_vtd",
	"Sala":             "dAen_yTK_tqz",
	"Skinnskatteberg":  "Nufj_vmt_VrH",
	"Surahammar":       "jfD3_Hdg_UhT",
	"Västerås":         "8deT_FRF_2SP",
	"Västra Götalands län": "zdoY_6u5_Krt",
	"Ale":              "17Ug_Btv_mBr",
	"Alingsås":         "UQ75_1eU_jaC",
	"Bengtsfors":       "hejM_Jct_XJk",
	"Bollebygd":        "ypAQ_vTD_KLU",
	"Borås":            "TpRZ_bFL_jhL",
	"Dals-Ed":          "NMc9_oEm_yxy",
	"Essunga":          "ZzEA_2Fg_Pt2",
	"Falköping":        "ZySF_gif_zE4",
	"Färgelanda":       "kCHb_icw_W5E",
	"Grästorp":         "ZNZy_Hh5_gSW",
	"Gullspång":        "roiB_uVV_4Cj",
	"Göteborg":         "PVZL_BQT_XtL",
	"Götene":           "txzq_PQY_FGi",
	"Herrljunga":       "J116_VFs_cg6",
	"Hjo":              "YbFS_34r_K2v",
	"Härryda":          "dzWW_R3G_6Eh",
	"Karlsborg":        "e413_94L_hdh",
	"Kungälv":          "ZkZf_HbK_Mcr",
	"Lerum":            "yHV7_2Y6_zQx",
	"Lidköping":        "FN1Y_asc_D8y",
	"Lilla Edet":       "YQcE_SNB_Tv3",
	"Lysekil":          "z2cX_rjC_zFo",
	"Mariestad":        "Lzpu_thX_Wpa",
	"Mark":             "7HAb_9or_eFM",
	"Mellerud":         "tt1B_7rH_vhG",
	"Munkedal":         "96Dh_3sQ_RFb",
	"Mölndal":          "mc45_ki9_Bv3",
	"Orust":            "tmAp_ykH_N6k",
	"Partille":         "CCiR_sXa_BVW",
	"Skara":            "k1SK_gxg_dW4",
	"Skövde":           "fqAy_4ji_Lz2",
	"Sotenäs":          "aKkp_sEX_cVM",
	"Stenungsund":      "wHrG_FBH_hoD",
	"Strömstad":        "PAxT_FLT_3Kq",
	"Svenljunga":       "rZWC_pXf_ySZ",
	"Tanum":            "qffn_qY4_DLk",
	"Tibro":            "aLFZ_NHw_atB",
	"Tidaholm":         "Zsf5_vpP_Bs4",
	"Tjörn":            "TbL3_HmF_gnx",
	"Tranemo":          "SEje_LdC_9qN",
	"Trollhättan":      "CSy8_41F_YvX",
	"Töreboda":         "a15F_gAH_pn6",
	"Uddevalla":        "xQc2_SzA_rHK",
	"Ulricehamn":       "an4a_8t2_Zpd",
	"Vara":             "fbHM_yhA_BqS",
	"Vänersborg":       "THif_q6H_MjG",
	"Vårgårda":         "NfFx_5jj_ogg",
	"Åmål":             "M1UC_Cnf_r7g",
	"Öckerö":           "Zjiv_rhk_oJK",
	"Örebro län":       "xTCk_nT5_Zjm",
	"Askersund":        "dbF7_Ecz_CWF",
	"Degerfors":        "pvzC_muj_rcq",
	"Hallsberg":        "Ak9V_rby_yYS",
	"Hällefors":        "sCbY_r36_xhs",
	"Karlskoga":        "wgJm_upX_z5W",
	"Kumla":            "viCA_36P_pQp",
	"Laxå":             "oYEQ_m8Q_unY",
	"Lekeberg":         "yaHU_E7z_YnE",
	"Lindesberg":       "JQE9_189_Ska",
	"Ljusnarsberg":     "eF2n_714_hSU",
	"Nora":             "WFXN_hsU_gmx",
	"Örebro":           "kuMn_feU_hXx",
	"Östergötlands län": "oLT3_Q9p_3nn",
	"Boxholm":          "e5LB_m9V_TnT",
	"Finspång":         "dMFe_J6W_iJv",
	"Kinda":            "U4XJ_hYF_FBA",
	"Linköping":        "bm2x_1mr_Qhx",
	"Mjölby":           "stqv_JGB_x8A",
	"Motala":           "E1MC_1uG_phm",
	"Norrköping":       "SYty_Yho_JAF",
	"Söderköping":      "Pcv9_yYh_Uw8",
	"Vadstena":         "VcCU_Y86_eKU",
	"Valdemarsvik":     "Sb3D_iGB_aXu",
	"Ydre":             "vRRz_nLT_vYv",
	"Åtvidaberg":       "bFWo_FRJ_x2T",
	"Ödeshög":          "Fu8g_29u_3xF",
}

// NormalizedMunicipalityMap innehåller normaliserade kommunnamn för snabbare sökning
var NormalizedMunicipalityMap map[string]string

func init() {
	// Initialisera den normaliserade kartan
	NormalizedMunicipalityMap = make(map[string]string)
	for k, v := range MunicipalityMap {
		normalized := normalizeString(k)
		NormalizedMunicipalityMap[normalized] = v
	}
}

// normalizeString normaliserar en sträng för jämförelse
func normalizeString(s string) string {
	if s == "" {
		return s
	}
	
	// Konvertera till lowercase först
	s = strings.ToLower(s)
	
	// Ersätt svenska tecken om de finns
	if strings.ContainsAny(s, "åäöéèëü") {
		replacements := []struct {
			old string
			new string
		}{
			{"å", "a"},
			{"ä", "a"},
			{"ö", "o"},
			{"é", "e"},
			{"è", "e"},
			{"ë", "e"},
			{"ü", "u"},
		}
		
		for _, r := range replacements {
			if strings.Contains(s, r.old) {
				s = strings.ReplaceAll(s, r.old, r.new)
			}
		}
	}
	
	// Behåll bara bokstäver och siffror
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// GetMunicipalityID returns the ID for a given municipality name
func GetMunicipalityID(name string) string {
	if name == "" {
		return ""
	}
	
	log.Printf("🔍 Söker efter kommun: '%s'", name)
	
	// Först, försök med direkt matchning mot originalnycklarna
	if id, exists := MunicipalityMap[name]; exists {
		log.Printf("✅ Hittade exakt matchning för kommun: %s -> %s", name, id)
		return id
	}
	
	// Normalisera söksträngen
	searchName := normalizeString(name)
	log.Printf("🔄 Normaliserad sökterm: '%s' -> '%s'", name, searchName)
	
	// Försök med normaliserad exakt matchning
	if id, exists := NormalizedMunicipalityMap[searchName]; exists {
		log.Printf("✅ Hittade normaliserad matchning för kommun: %s -> %s", name, id)
		return id
	}
	
	// Om ingen exakt matchning, försök med delmatchning
	for normalizedKey, id := range NormalizedMunicipalityMap {
		if strings.Contains(normalizedKey, searchName) ||
			strings.Contains(searchName, normalizedKey) {
			log.Printf("✅ Hittade delmatchning för kommun: %s -> %s (nyckel: %s)", name, id, normalizedKey)
			return id
		}
	}
	
	log.Printf("❌ Kunde inte hitta ID för kommun: %s (normaliserad: %s)", name, searchName)
	return ""
}

// GetMunicipalityName returnerar kommunnamnet för ett givet ID
func GetMunicipalityName(id string) string {
	for name, municipalityID := range MunicipalityMap {
		if municipalityID == id {
			return name
		}
	}
	return ""
}
