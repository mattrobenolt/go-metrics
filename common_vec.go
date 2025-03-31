package metrics

import "hash/maphash"

type commonVec struct {
	// either set or setvec will be non-nil
	set    *Set
	setvec *SetVec

	family      Ident
	partialTags []Label
	partialHash *maphash.Hash
}

func getCommonVecSet(s *Set, family string, labels []string) commonVec {
	return commonVec{
		set:         s,
		family:      MustIdent(family),
		partialTags: makeLabels(labels),
		partialHash: hashStart(family, labels...),
	}
}

func getCommonVecSetVec(sv *SetVec, family string, labels []string) commonVec {
	return commonVec{
		setvec:      sv,
		family:      MustIdent(family),
		partialTags: makeLabels(labels),
		partialHash: hashStart(family, labels...),
	}
}
