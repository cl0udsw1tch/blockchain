package transaction

type OpStack struct {
	items [][]byte
}

func (s *OpStack) Size() int{
	return len(s.items)
}

func (s *OpStack) Push(item []byte){
	s.items = append(s.items, item)
}

func (s *OpStack) Pop() []byte{
	if s.Size() == 0 {
		return nil
	}

	lastIndex := s.Size() - 1
	popped := s.items[lastIndex]
	s.items = s.items[:lastIndex]
	return popped
}

func (s *OpStack) Peek() []byte {
	
	if s.Size() == 0{
		return nil
	}

	return s.items[s.Size() - 1]
}

func (s *OpStack) IsEmpty() bool {

    return s.Size() == 0
}