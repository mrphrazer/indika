package blanket_emulator

import (
	uc "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)

const pagesize = 4096

type WorkingSet struct {
	mapped []uint64
	newest int
	oldest int
}

func NewWorkingSet(size int) *WorkingSet {
	res := new(WorkingSet)
	res.mapped = make([]uint64, size)
	res.newest = -1
	res.oldest = -1
	return res
}

func (s *WorkingSet) Map(addr, size uint64, mu uc.Unicorn) error {
	alignment := (addr % pagesize)
	base_addr := addr - alignment
	err := mu.MemMap(base_addr, uint64(pagesize))
	if err != nil {
		return err
	}
	err = mu.MemWrite(base_addr, GetMem(base_addr, pagesize))
	if err != nil {
		return err
	}
	s.StoreInWorkingSet(base_addr, mu)
	if addr+size > base_addr+pagesize { //sometimes we might need to map 2 pages
		s.Map(base_addr+pagesize, 1, mu) //map next pages as well
	}
	return nil
}

func (s *WorkingSet) StoreInWorkingSet(addr uint64, mu uc.Unicorn) error {
	if s.newest == -1 {
		s.mapped[0] = addr
		s.oldest = 0
		s.newest = 0
	}
	s.newest = (s.newest + 1) % len(s.mapped)
	if s.newest == s.oldest { // unmap old page
		addr_to_unmap := s.mapped[s.oldest]
		err := mu.MemUnmap(addr_to_unmap, pagesize)
		if err != nil {
			return err
		}
	}
	s.oldest = (s.oldest + 1) % len(s.mapped)
	s.mapped[s.newest] = addr
	return nil
}

func (s *WorkingSet) Clear(mu uc.Unicorn) error {
	if s.newest == -1 {
		return nil
	}
	for i := s.oldest; i != s.newest; i = (i + 1) % len(s.mapped) {
		err := mu.MemUnmap(s.mapped[i], pagesize)
		if err != nil {
			return err
		}
	}
	return nil
}
