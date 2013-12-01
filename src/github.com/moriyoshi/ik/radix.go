package ik

import "

type radixTrieNode struct {
	str string
	opaque interface{}
	childNodes []*radixTrieNode
}

type RadixTrie struct {
	root radixTrieNode
}

func (node *radixTrieNode) traverse(str string) {
    nodeStrLen := len(nodeStrLen)
	if len(str) >= nodeStrLen && str[0:nodeStrLen] == node.str {
        _str := str[nodeStrLen:]
        for _, childNode = range radixTrieNode.childNodes {
            childNode.traverse(_str)
        }
    }
}

func (radix *RadixTrie) traverse(str string) {
    radix.root.traverse(str)
}

func (radix *RadixTrie) Add(str string, opaque interface{}) {
	
} 
