package main

const MAX_DEGREE int = 4 // MAX_DEGREE = 4 means that in one node at most 3 elements can be there and at most 4 children of that node
const min_block_size int = (MAX_DEGREE / 2) - 1

type Node struct {
	block_size    int
	children_size int
	block         [MAX_DEGREE]int
	children      [MAX_DEGREE + 1]*Node
}

func binary_index(arr []int, low, high, x int) (int, bool) {
	l, r := low, high
	mid := 0
	for l < r {
		mid = (l + r) / 2
		if arr[mid] == x {
			return mid, true
		} else if arr[mid] > x {
			r = mid
		} else {
			l = mid + 1
		}
	}
	return l, false
}

func fix_children_size(node *Node) {
	node.children_size = 0
	for i := 0; i < MAX_DEGREE+1; i++ { // This part of code is HIGHLY VARIANT and SENSITIVE... Small changes can break the code
		if node.children[i] != nil {
			node.children_size = i + 1
		}
	}
}

func insert_into_slice(node *Node, x, insert_ind int, new_node *Node, is_left_child_pref bool) {
	/*
		ASSUMPTION:
			The input slice has NECESSARILY an extra space at the index `node.block_size`
	*/
	var i int
	for i = node.block_size; i > insert_ind; i-- {
		node.block[i] = node.block[i-1]
	}
	node.block[insert_ind] = x

	limit := insert_ind + 1
	if is_left_child_pref {
		limit = insert_ind
	}

	for i = node.children_size; i > limit; i-- {
		node.children[i] = node.children[i-1]
	}
	node.children[limit] = new_node
	node.block_size += 1
	fix_children_size(node)
}

func split(node *Node) (int, *Node) {

	var mid int
	var new_node Node
	var i int = 0

	mid = MAX_DEGREE / 2
	for i = mid + 1; i < node.block_size; i++ {
		new_node.block[i-mid-1] = node.block[i]
		node.block[i] = 0
	}
	new_node.block_size = node.block_size - mid - 1
	node.block_size = mid

	for i = mid + 1; i < node.children_size; i++ {
		new_node.children[i-mid-1] = node.children[i]
		node.children[i] = nil
	}

	fix_children_size(node)
	fix_children_size(&new_node)

	push_to_top := node.block[mid]
	node.block[mid] = 0

	return push_to_top, &new_node
}

func insert_helper(node *Node, x int) (*Node, bool, int, *Node) {
	/*
		INPUT:
			1. Current root of the B-Tree
			2. `x` integer to insert in the B-Tree
		OUTPUT:
			1. *Node => The new root of the B-Tree
			2. bool => `is_overflow` -> True, if overflow occured while inserting in the B-Tree, (signalling splitting has occured)
			3. int => `y` -> The integer which the current needs to accomodate since the lower layers are full.
			4. *Node => New right node created from splitting at bottom level (This needs to be adjusted in the current node)
	*/

	// Check if there are any children
	if node.children_size == 0 {
		// This is the leaf node
		ind, inArr := binary_index(node.block[:], 0, node.block_size, x)
		if inArr {
			return node, false, -1, nil
		}

		insert_into_slice(node, x, ind, nil, false)

		if node.block_size == MAX_DEGREE {
			// Overflow has occured in the node
			push_to_top, new_node := split(node)
			return node, true, push_to_top, new_node
		}

		return node, false, -1, nil

	} else {
		// Now, we need to find the correct child to do recursion on
		ind, inArr := binary_index(node.block[:], 0, node.block_size, x)
		if inArr {
			return node, false, -1, nil
		}

		var is_overflow bool
		var pushed_from_bottom int
		var new_node *Node
		_, is_overflow, pushed_from_bottom, new_node = insert_helper(node.children[ind], x)

		if !is_overflow {
			return node, false, -1, nil
		}

		ind, _ = binary_index(node.block[:], 0, node.block_size, pushed_from_bottom)
		insert_into_slice(node, pushed_from_bottom, ind, new_node, false)

		if node.block_size == MAX_DEGREE {
			// Overflow has occured in the node
			push_to_top, new_node := split(node)
			return node, true, push_to_top, new_node
		}

		return node, false, -1, nil
	}
}

func delete_from_slice(node *Node, ind int, is_left_child_pref bool) (int, *Node) {

	if ind >= node.block_size {
		return -1, nil
	}

	var j int
	temp2 := node.block[ind]
	for j = ind; j < node.block_size; j++ {
		node.block[j] = node.block[j+1]
	}
	j = ind + 1
	if is_left_child_pref {
		j = ind
	}
	temp := node.children[j]
	for j < node.children_size {
		node.children[j] = node.children[j+1]
		j++
	}
	node.block_size -= 1

	fix_children_size(node)

	return temp2, temp
}

func merge_helper(left_node, right_node *Node) {

	i := left_node.block_size
	j := 0
	for i < MAX_DEGREE-1 && j < right_node.block_size {
		left_node.block[i] = right_node.block[j]
		i++
		j++
	}

	i = left_node.children_size
	j = 1
	for i < MAX_DEGREE && j < right_node.children_size {
		left_node.children[i] = right_node.children[j]
		i++
		j++
	}
	left_node.block_size += right_node.block_size
	fix_children_size(left_node)
	fix_children_size(right_node)
}

func merge(node *Node, ind int) {
	/*
		4 Possible Cases:
			1. The child node has more than `min_block_size` elements.
			2. The sibling node on right has more than `min_block_size` elems.
			3. The sibling node on right does not have more than `min_block_size` elems. But, the left does.
			4. Neither have more than `min_block_size` elems.
				4.1. Left Sibling exists
				4.2. Right Sibling exists
		[Other sibling nodes except for just right and just left are not useful (in this scenario)]
	*/

	if node.children[ind].block_size >= min_block_size {
		return
	}
	if (ind+1 < node.children_size) && (node.children[ind+1].block_size > min_block_size) {
		// Right sibling exists and has more than `min_block_size` elements
		num, child := delete_from_slice(node.children[ind+1], 0, true)

		temp := node.block[ind]
		node.block[ind] = num

		insert_ind, _ := binary_index(node.children[ind].block[:], 0, node.children[ind].block_size, temp)
		insert_into_slice(node.children[ind], temp, insert_ind, child, false)

		return
	}
	if (ind-1 > -1) && (node.children[ind-1].block_size > min_block_size) {
		// Left sibling exists and has more than `min_block_size` elements
		num, child := delete_from_slice(node.children[ind-1], node.children[ind-1].block_size-1, false)

		temp := node.block[ind-1]
		node.block[ind-1] = num

		insert_ind, _ := binary_index(node.children[ind].block[:], 0, node.children[ind].block_size, temp)
		insert_into_slice(node.children[ind], temp, insert_ind, child, true)

		return
	}
	// Neither left nor right sibling have extra elems to spare. So, we possibly merge them (if they exist)
	if ind-1 > -1 {
		// Left Sibling exists
		insert_into_slice(node.children[ind], node.block[ind-1], 0, nil, true)
		merge_helper(node.children[ind-1], node.children[ind])
		delete_from_slice(node, ind-1, false)

	} else {
		// Right Sibling exists
		insert_into_slice(node.children[ind+1], node.block[ind], 0, nil, true)
		merge_helper(node.children[ind], node.children[ind+1])
		delete_from_slice(node, ind, false)
	}
}

func find_leftmost(node *Node) int {
	if node.children_size == 0 {
		return node.block[0]
	}
	return find_leftmost(node.children[0])
}

func delete_helper(node *Node, x int) int {
	/*
		INPUT:
			1. The root of the B-Tree in which the element could be present
			2. The element which needs to be deleted from the B-Tree
		OUTPUT:
			1. An integer which represents a code, which is used for internal function use to tell what case are we working with.
				-> -1 = The element was not found
				->  0 = The element was found and deleted wih no problems
				->  1 = The element is in a leaf node (Just below the current node)
				->  2 = The element is in an internal node (Just below the current node) [NEVER ACTUALLY IS RETURNED] [Is Implicit]
				->  3 = The deletion has occured but now the node below has less number of elems and thus needs to be merged with sibling.
	*/
	if node == nil {
		return -1
	}
	ind, inArr := binary_index(node.block[:], 0, node.block_size, x)

	if inArr { // This node contains the element we want to delete
		if node.children_size == 0 {
			del_ind, _ := binary_index(node.block[:], 0, node.block_size, x)
			delete_from_slice(node, del_ind, false)
			return 1 // This is a leaf node
		}
		// This is an internal node
		replace_num := find_leftmost(node.children[ind+1])
		node.block[ind] = replace_num
		delete_helper(node.children[ind+1], replace_num)
		if node.children[ind+1].block_size < min_block_size {
			merge(node, ind+1)
		}
		return 3
	}

	// The element maybe in one of the leaf nodes of the current node
	code := delete_helper(node.children[ind], x)

	if code == -1 {
		// Element was not found in the children
		return -1

	} else if code == 0 {
		// The element was found and deleted wih no problems
		return 0

	} else if code == 1 {
		// Element WAS in the leaf node (which is just below this node)
		merge(node, ind)
		if node.block_size < min_block_size {
			return 3
		}
		return 0

	} else if code == 3 {
		merge(node, ind)
		if node.block_size < min_block_size {
			return 3
		}
		return 0

	}

	return 0
}

func Search(node *Node, x int) *int {
	if node == nil {
		return nil
	}
	ind, inArr := binary_index(node.block[:], 0, node.block_size, x)
	if inArr {
		return &node.block[ind]
	}
	return Search(node.children[ind], x)
}

func Insert(root *Node, x int) *Node {

	var new_root, new_node *Node
	var y int
	var is_overflow bool

	new_root, is_overflow, y, new_node = insert_helper(root, x)
	if !is_overflow {
		return new_root
	}

	var node Node
	node.block_size = 1
	node.block[0] = y
	node.children[0] = new_root
	node.children[1] = new_node
	node.children_size = 0
	for i := 0; i < MAX_DEGREE+1; i++ {
		if node.children[i] != nil {
			node.children_size = i + 1
		}
	}
	return &node
}

func Delete(node *Node, x int) *Node {

	delete_helper(node, x)
	if node.block_size < min_block_size {
		return node.children[0]
	}
	return node
}

func Update(node *Node, x, y int) *Node {
	node = Delete(node, x)
	node = Insert(node, y)
	return node
}
